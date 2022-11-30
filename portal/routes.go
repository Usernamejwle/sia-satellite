package portal

import (
	"bytes"
	"net/http"
	"net/mail"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/julienschmidt/httprouter"
)

const (
	// verifyTemplate contains the text send by email when a
	// new user account is being created.
	verifyTemplate = `
		<!-- template.html -->
		<!DOCTYPE html>
		<html>
		<body>
    	<h2>Please Verify Your Email Address</h2>
	    <p>Click on the following link to complete your account registration. This link is valid within the next 24 hours.</p>
	    <p><a href="{{.Path}}?token={{.Token}}">{{.Path}}?token={{.Token}}</a></p>
		</body>
		</html>
	`

	// resetTemplate contains the text send by email when a
	// user wants to reset their password.
	resetTemplate = `
		<!-- template.html -->
		<!DOCTYPE html>
		<html>
		<body>
    	<h2>Reset Your Password</h2>
	    <p>Click on the following link to enter a new password. This link is valid within the next 24 hours.</p>
	    <p><a href="{{.Path}}?token={{.Token}}">{{.Path}}?token={{.Token}}</a></p>
		</body>
		</html>
	`
)

// authLink holds the parts of an authentication link.
type authLink struct {
	Path  string
	Token string
}

type (
	// authRequest holds the body of an /authme POST request.
	authRequest struct {
		Email    string `json: "email"`
		Password string `json: "password"`
	}
)

// checkEmail is a helper function that validates an email address.
// If the email address is valid, it is returned in lowercase.
func checkEmail(address string) (string, Error) {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return "", Error{
			Code: httpErrorEmailInvalid,
			Message: "the email address is invalid",
		}
	}
	if len(address) > 48 {
		return "", Error{
			Code: httpErrorEmailTooLong,
			Message: "the email address is too long",
		}
	}
	return strings.ToLower(address), Error{}
}

// checkPassword is a helper function that checks if the password
// complies with the rules.
func checkPassword(pwd string) Error {
	if len(pwd) < 8 {
		return Error{
			Code: httpErrorPasswordTooShort,
			Message: "the password is too short",
		}
	}
	if len(pwd) > 255 {
		return Error{
			Code: httpErrorPasswordTooLong,
			Message: "the password is too long",
		}
	}
	var smalls, caps, digits, specials bool
	for _, c := range pwd {
		switch {
			case unicode.IsNumber(c):
				digits = true
			case unicode.IsLower(c):
				smalls = true
			case unicode.IsUpper(c):
				caps = true
			case unicode.IsPunct(c) || unicode.IsSymbol(c):
				specials = true
			default:
		}
	}
	if !smalls || !caps || !digits || !specials {
		return Error{
			Code: httpErrorPasswordNotCompliant,
			Message: "insecure password",
		}
	}
	return Error{}
}

// authHandlerPOST handles the POST /auth requests.
func (api *portalAPI) authHandlerPOST(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	dec, decErr := prepareDecoder(w, req)
	if decErr != nil {
		return
	}

	var auth authRequest
	err, code := api.handleDecodeError(w, dec.Decode(&auth))
	if code != http.StatusOK {
		writeError(w, err, code)
		return
	}

	// TODO implement auth code.

	writeSuccess(w)
}

// registerHandlerPOST handles the POST /register requests.
func (api *portalAPI) registerHandlerPOST(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Decode request body.
	dec, decErr := prepareDecoder(w, req)
	if decErr != nil {
		return
	}

	var reg authRequest
	err, code := api.handleDecodeError(w, dec.Decode(&reg))
	if code != http.StatusOK {
		writeError(w, err, code)
		return
	}

	// Check request fields for validity.	
	email, err := checkEmail(reg.Email)
	if err.Code != httpErrorNone {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	password := reg.Password
	if err := checkPassword(password); err.Code != httpErrorNone {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	// Check if the email address is already registered.
	registeredAndVerified := false
	count, cErr := api.portal.countEmails(email)
	if cErr != nil {
		api.portal.log.Printf("ERROR: error querying database: %v\n", cErr)
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "internal error",
			}, http.StatusInternalServerError)
		return
	}
	if count > 0 {
		// Check if the account is verified already.
		var passwordOK bool
		registeredAndVerified, passwordOK, cErr = api.portal.isVerified(email, password)
		if cErr != nil {
			api.portal.log.Printf("ERROR: error querying database: %v\n", cErr)
			writeError(w,
				Error{
					Code: httpErrorInternal,
					Message: "internal error",
				}, http.StatusInternalServerError)
			return
		}

		// Wrong password.
		if !passwordOK {
			// Check and update login stats.
			if cErr := api.portal.checkAndUpdateFailedLogins(req.RemoteAddr); cErr != nil {
				writeError(w,
					Error{
						Code: httpErrorTooManyRequests,
						Message: "too many failed login attempts",
					}, http.StatusTooManyRequests)
				return
			}

			writeError(w,
				Error{
					Code: httpErrorWrongCredentials,
					Message: "invalid combination of email and password",
				}, http.StatusBadRequest)
			return
		}

		// This account is completely registered.
		if registeredAndVerified {
			writeError(w,
				Error{
					Code: httpErrorEmailUsed,
					Message: "email already registered",
				}, http.StatusBadRequest)
			return
		}
	}

	// Send verification link by email.
	if !api.sendVerificationLinkByMail(w, req, email) {
		return
	}

	if !registeredAndVerified {
		// Create a new account.
		if cErr := api.portal.updateAccount(email, password, false); cErr != nil {
			api.portal.log.Printf("ERROR: error querying database: %v\n", cErr)
			writeError(w,
				Error{
					Code: httpErrorInternal,
					Message: "internal error",
				}, http.StatusInternalServerError)
			return
		}

		writeSuccess(w)
	} else {
		// Report that the account is unverified.
		writeError(w, Error{
			Code: httpErrorNone,
			Message: "account not verified yet",
		}, http.StatusOK)
		return
	}
}

// sendVerificationLinkByMail is a wrapper function for sending a
// verification link by email.
func (api *portalAPI) sendVerificationLinkByMail(w http.ResponseWriter, req *http.Request, email string) bool {
	// Check and update stats.
	if err := api.portal.checkAndUpdateVerifications(req.RemoteAddr); err != nil {
		writeError(w,
			Error{
				Code: httpErrorTooManyRequests,
				Message: "too many verification requests",
			}, http.StatusTooManyRequests)
		return false
	}

	// Generate a verification link.
	token := api.portal.generateToken(verifyPrefix, email, time.Now().Add(24 * time.Hour))
	path := req.Header["Referer"]
	if len(path) == 0 {
		api.portal.log.Printf("ERROR: unable to fetch referer URL")
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "unable to fetch referer URL",
			}, http.StatusInternalServerError)
		return false
	}
	link := authLink{
		Path:  path[0],
		Token: token,
	}

	// Generate email body.
	t := template.New("verify")
	t, err := t.Parse(verifyTemplate)
	if err != nil {
		api.portal.log.Printf("ERROR: unable to parse HTML template: %v\n", err)
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "unable to send verification link",
			}, http.StatusInternalServerError)
		return false
	}
	var b bytes.Buffer
	t.Execute(&b, link)

	// Send verification link by email.
	err = api.portal.ms.SendMail("Sia Satellite", email, "Action Required", &b)
	if err != nil {
		api.portal.log.Printf("ERROR: unable to send verification link: %v\n", err)
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "unable to send verification link",
			}, http.StatusInternalServerError)
		return false
	}

	return true
}

// sendPasswordResetLinkByMail is a wrapper function for sending a
// password reset link by email.
func (api *portalAPI) sendPasswordResetLinkByMail(w http.ResponseWriter, req *http.Request, email string) bool {
	// Generate a password reset link.
	token := api.portal.generateToken(resetPrefix, email, time.Now().Add(24 * time.Hour))
	path := req.Header["Referer"]
	if len(path) == 0 {
		api.portal.log.Printf("ERROR: unable to fetch referer URL")
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "unable to fetch referer URL",
			}, http.StatusInternalServerError)
		return false
	}
	link := authLink{
		Path:  path[0],
		Token: token,
	}

	// Generate email body.
	t := template.New("reset")
	t, err := t.Parse(resetTemplate)
	if err != nil {
		api.portal.log.Printf("ERROR: unable to parse HTML template: %v\n", err)
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "unable to send password reset link",
			}, http.StatusInternalServerError)
		return false
	}
	var b bytes.Buffer
	t.Execute(&b, link)

	// Send password reset link by email.
	err = api.portal.ms.SendMail("Sia Satellite", email, "Reset Your Password", &b)
	if err != nil {
		api.portal.log.Printf("ERROR: unable to send password reset link: %v\n", err)
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "unable to send password reset link",
			}, http.StatusInternalServerError)
		return false
	}

	return true
}

// registerResendHandlerPOST handles the POST /register/resend requests.
func (api *portalAPI) registerResendHandlerPOST(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Decode request body.
	dec, decErr := prepareDecoder(w, req)
	if decErr != nil {
		return
	}

	var data struct {
		Email string `json: "email"`
	}
	err, code := api.handleDecodeError(w, dec.Decode(&data))
	if code != http.StatusOK {
		writeError(w, err, code)
		return
	}

	// Send verification link by email.
	if !api.sendVerificationLinkByMail(w, req, data.Email) {
		return
	}

	writeSuccess(w)
}

// tokenHandlerPOST handles the POST /token requests.
func (api *portalAPI) tokenHandlerPOST(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Decode request body.
	dec, decErr := prepareDecoder(w, req)
	if decErr != nil {
		return
	}

	var data struct {
		Token string `json: "token"`
	}
	err, code := api.handleDecodeError(w, dec.Decode(&data))
	if code != http.StatusOK {
		writeError(w, err, code)
		return
	}

	// Decode the token.
	prefix, email, expires, tErr := api.portal.decodeToken(data.Token)
	if tErr != nil {
		writeError(w,
			Error{
				Code: httpErrorTokenInvalid,
				Message: "unable to decode token",
			}, http.StatusBadRequest)
		return
	}

	switch prefix {
	case verifyPrefix:
		if expires.Before(time.Now()) {
			writeError(w,
			Error{
				Code: httpErrorTokenExpired,
				Message: "link already expired",
			}, http.StatusBadRequest)
			return
		}
		err := api.portal.updateAccount(email, "", true)
		if err != nil {
			writeError(w,
				Error{
					Code: httpErrorInternal,
					Message: "unable to verify account",
				}, http.StatusInternalServerError)
			return
		}

	case resetPrefix:
		// TODO

	default:
		writeError(w,
			Error{
				Code: httpErrorTokenInvalid,
				Message: "prefix not supported",
			}, http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

// resetHandlerPOST handles the POST /reset requests.
func (api *portalAPI) resetHandlerPOST(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Check and update password reset stats.
	if cErr := api.portal.checkAndUpdatePasswordResets(req.RemoteAddr); cErr != nil {
		writeError(w,
			Error{
				Code: httpErrorTooManyRequests,
				Message: "too many password reset requests",
			}, http.StatusTooManyRequests)
		return
	}

	// Decode request body.
	dec, decErr := prepareDecoder(w, req)
	if decErr != nil {
		return
	}

	var data struct {
		Email string `json: "email"`
	}
	err, code := api.handleDecodeError(w, dec.Decode(&data))
	if code != http.StatusOK {
		writeError(w, err, code)
		return
	}

	// Check if such account exists.
	count, cErr := api.portal.countEmails(data.Email)
	if cErr != nil {
		api.portal.log.Printf("ERROR: error querying database: %v\n", cErr)
		writeError(w,
			Error{
				Code: httpErrorInternal,
				Message: "internal error",
			}, http.StatusInternalServerError)
		return
	}

	if count == 0 {
		// Do not return an error. Otherwise we would give a potential
		// attacker a hint.
		writeSuccess(w)
		return
	}

	// Send password reset link by email.
	if !api.sendPasswordResetLinkByMail(w, req, data.Email) {
		return
	}

	writeSuccess(w)
}
