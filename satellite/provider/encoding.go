package provider

import (
	rhpv2 "go.sia.tech/core/rhp/v2"
	"go.sia.tech/core/types"
	"go.sia.tech/siad/crypto"
)

var (
	// Handshake specifier.
	loopEnterSpecifier = types.NewSpecifier("LoopEnter")

	// RPC ciphers.
	cipherChaCha20Poly1305 = types.NewSpecifier("ChaCha20Poly1305")
	cipherNoOverlap        = types.NewSpecifier("NoOverlap")
)

// Handshake objects.
type (
	loopKeyExchangeRequest struct {
		Specifier types.Specifier
		PublicKey [32]byte
		Ciphers   []types.Specifier
	}

	loopKeyExchangeResponse struct {
		PublicKey [32]byte
		Signature types.Signature
		Cipher    types.Specifier
	}
)

// EncodeTo implements types.ProtocolObject.
func (r *loopKeyExchangeRequest) EncodeTo(e *types.Encoder) {
	// Nothing to do here.
}

// DecodeFrom implements types.ProtocolObject.
func (r *loopKeyExchangeRequest) DecodeFrom(d *types.Decoder) {
	r.Specifier.DecodeFrom(d)
	d.Read(r.PublicKey[:])
	r.Ciphers = make([]types.Specifier, d.ReadPrefix())
	for i := range r.Ciphers {
		r.Ciphers[i].DecodeFrom(d)
	}
}

// EncodeTo implements types.ProtocolObject.
func (r *loopKeyExchangeResponse) EncodeTo(e *types.Encoder) {
	e.Write(r.PublicKey[:])
	e.WriteBytes(r.Signature[:])
	r.Cipher.EncodeTo(e)
}

// DecodeFrom implements types.ProtocolObject.
func (r *loopKeyExchangeResponse) DecodeFrom(d *types.Decoder) {
	// Nothing to do here.
}

// rpcError is the generic error transferred in an RPC.
type rpcError struct {
	Type        types.Specifier
	Data        []byte
	Description string
}

// EncodeTo implements types.ProtocolObject.
func (re *rpcError) EncodeTo(e *types.Encoder) {
	e.Write(re.Type[:])
	e.WriteBytes(re.Data)
	e.WriteString(re.Description)
}

// DecodeFrom implements types.ProtocolObject.
func (re *rpcError) DecodeFrom(d *types.Decoder) {
	// Nothing to do here.
}

// rpcResponse if a helper type for encoding and decoding RPC response
// messages, which can represent either valid data or an error.
type rpcResponse struct {
	err  *rpcError
	data requestBody
}

// EncodeTo implements types.ProtocolObject.
func (resp *rpcResponse) EncodeTo(e *types.Encoder) {
	e.WriteBool(resp.err != nil)
	if resp.err != nil {
		resp.err.EncodeTo(e)
		return
	}
	if resp.data != nil {
		resp.data.EncodeTo(e)
	}
}

// DecodeFrom implements types.ProtocolObject.
func (resp *rpcResponse) DecodeFrom(d *types.Decoder) {
	// Nothing to do here.
}

// requestBody is the common interface type for the renter requests.
type requestBody interface {
	DecodeFrom(d *types.Decoder)
	EncodeTo(e *types.Encoder)
}

// requestRequest is used when the renter requests the list of their
// active contracts.
type requestRequest struct {
	PubKey    crypto.PublicKey
	Signature types.Signature
}

// DecodeFrom implements requestBody.
func (rr *requestRequest) DecodeFrom(d *types.Decoder) {
	copy(rr.PubKey[:], d.ReadBytes())
	rr.Signature.DecodeFrom(d)
}

// EncodeTo implements requestBody.
func (rr *requestRequest) EncodeTo(e *types.Encoder) {
	e.WriteBytes(rr.PubKey[:])
}

// formRequest is used when the renter requests forming contracts with
// the hosts.
type formRequest struct {
	PubKey      crypto.PublicKey
	Hosts       uint64
	Period      uint64
	RenewWindow uint64

	Storage  uint64
	Upload   uint64
	Download uint64

	MinShards   uint64
	TotalShards uint64

	MaxRPCPrice          types.Currency
	MaxContractPrice     types.Currency
	MaxDownloadPrice     types.Currency
	MaxUploadPrice       types.Currency
	MaxStoragePrice      types.Currency
	MaxSectorAccessPrice types.Currency
	MinMaxCollateral     types.Currency
	BlockHeightLeeway    uint64

	Signature types.Signature
}

// DecodeFrom implements requestBody.
func (fr *formRequest) DecodeFrom(d *types.Decoder) {
	copy(fr.PubKey[:], d.ReadBytes())
	fr.Hosts = d.ReadUint64()
	fr.Period = d.ReadUint64()
	fr.RenewWindow = d.ReadUint64()
	fr.Storage = d.ReadUint64()
	fr.Upload = d.ReadUint64()
	fr.Download = d.ReadUint64()
	fr.MinShards = d.ReadUint64()
	fr.TotalShards = d.ReadUint64()
	fr.MaxRPCPrice.DecodeFrom(d)
	fr.MaxContractPrice.DecodeFrom(d)
	fr.MaxDownloadPrice.DecodeFrom(d)
	fr.MaxUploadPrice.DecodeFrom(d)
	fr.MaxStoragePrice.DecodeFrom(d)
	fr.MaxSectorAccessPrice.DecodeFrom(d)
	fr.MinMaxCollateral.DecodeFrom(d)
	fr.BlockHeightLeeway = d.ReadUint64()
	fr.Signature.DecodeFrom(d)
}

// EncodeTo implements requestBody.
func (fr *formRequest) EncodeTo(e *types.Encoder) {
	e.WriteBytes(fr.PubKey[:])
	e.WriteUint64(fr.Hosts)
	e.WriteUint64(fr.Period)
	e.WriteUint64(fr.RenewWindow)
	e.WriteUint64(fr.Storage)
	e.WriteUint64(fr.Upload)
	e.WriteUint64(fr.Download)
	e.WriteUint64(fr.MinShards)
	e.WriteUint64(fr.TotalShards)
	fr.MaxRPCPrice.EncodeTo(e)
	fr.MaxContractPrice.EncodeTo(e)
	fr.MaxDownloadPrice.EncodeTo(e)
	fr.MaxUploadPrice.EncodeTo(e)
	fr.MaxStoragePrice.EncodeTo(e)
	fr.MaxSectorAccessPrice.EncodeTo(e)
	fr.MinMaxCollateral.EncodeTo(e)
	e.WriteUint64(fr.BlockHeightLeeway)
}

// renewRequest is used when the renter requests contract renewals.
type renewRequest struct {
	PubKey      crypto.PublicKey
	Contracts   []types.FileContractID
	Period      uint64
	RenewWindow uint64

	Storage  uint64
	Upload   uint64
	Download uint64

	MinShards   uint64
	TotalShards uint64

	MaxRPCPrice          types.Currency
	MaxContractPrice     types.Currency
	MaxDownloadPrice     types.Currency
	MaxUploadPrice       types.Currency
	MaxStoragePrice      types.Currency
	MaxSectorAccessPrice types.Currency
	MinMaxCollateral     types.Currency
	BlockHeightLeeway    uint64

	Signature types.Signature
}

// DecodeFrom implements requestBody.
func (rr *renewRequest) DecodeFrom(d *types.Decoder) {
	copy(rr.PubKey[:], d.ReadBytes())
	numContracts := int(d.ReadUint64())
	rr.Contracts = make([]types.FileContractID, numContracts)
	for i := 0; i < numContracts; i++ {
		copy(rr.Contracts[i][:], d.ReadBytes())
	}
	rr.Period = d.ReadUint64()
	rr.RenewWindow = d.ReadUint64()
	rr.Storage = d.ReadUint64()
	rr.Upload = d.ReadUint64()
	rr.Download = d.ReadUint64()
	rr.MinShards = d.ReadUint64()
	rr.TotalShards = d.ReadUint64()
	rr.MaxRPCPrice.DecodeFrom(d)
	rr.MaxContractPrice.DecodeFrom(d)
	rr.MaxDownloadPrice.DecodeFrom(d)
	rr.MaxUploadPrice.DecodeFrom(d)
	rr.MaxStoragePrice.DecodeFrom(d)
	rr.MaxSectorAccessPrice.DecodeFrom(d)
	rr.MinMaxCollateral.DecodeFrom(d)
	rr.BlockHeightLeeway = d.ReadUint64()
	rr.Signature.DecodeFrom(d)
}

// EncodeTo implements requestBody.
func (rr *renewRequest) EncodeTo(e *types.Encoder) {
	e.WriteBytes(rr.PubKey[:])
	e.WriteUint64(uint64(len(rr.Contracts)))
	for _, id := range rr.Contracts {
		e.WriteBytes(id[:])
	}
	e.WriteUint64(rr.Period)
	e.WriteUint64(rr.RenewWindow)
	e.WriteUint64(rr.Storage)
	e.WriteUint64(rr.Upload)
	e.WriteUint64(rr.Download)
	e.WriteUint64(rr.MinShards)
	e.WriteUint64(rr.TotalShards)
	rr.MaxRPCPrice.EncodeTo(e)
	rr.MaxContractPrice.EncodeTo(e)
	rr.MaxDownloadPrice.EncodeTo(e)
	rr.MaxUploadPrice.EncodeTo(e)
	rr.MaxStoragePrice.EncodeTo(e)
	rr.MaxSectorAccessPrice.EncodeTo(e)
	rr.MinMaxCollateral.EncodeTo(e)
	e.WriteUint64(rr.BlockHeightLeeway)
}

// updateRequest is used when the renter submits a new revision.
type updateRequest struct {
	PubKey      crypto.PublicKey
	Contract    rhpv2.ContractRevision
	Uploads     types.Currency
	Downloads   types.Currency
	FundAccount types.Currency

	Signature types.Signature
}

// DecodeFrom implements requestBody.
func (ur *updateRequest) DecodeFrom(d *types.Decoder) {
	copy(ur.PubKey[:], d.ReadBytes())
	ur.Contract.Revision.DecodeFrom(d)
	ur.Contract.Signatures[0].DecodeFrom(d)
	ur.Contract.Signatures[1].DecodeFrom(d)
	ur.Uploads.DecodeFrom(d)
	ur.Downloads.DecodeFrom(d)
	ur.FundAccount.DecodeFrom(d)
	ur.Signature.DecodeFrom(d)
}

// EncodeTo implements requestBody.
func (ur *updateRequest) EncodeTo(e *types.Encoder) {
	e.WriteBytes(ur.PubKey[:])
	ur.Contract.Revision.EncodeTo(e)
	ur.Contract.Signatures[0].EncodeTo(e)
	ur.Contract.Signatures[1].EncodeTo(e)
	ur.Uploads.EncodeTo(e)
	ur.Downloads.EncodeTo(e)
	ur.FundAccount.EncodeTo(e)
}

// extendedContract contains the contract and its metadata.
type extendedContract struct {
	contract            rhpv2.ContractRevision
	startHeight         uint64
	totalCost           types.Currency
	uploadSpending      types.Currency
	downloadSpending    types.Currency
	fundAccountSpending types.Currency
	renewedFrom         types.FileContractID
}

// extendedContractSet is a collection of extendedContracts.
type extendedContractSet struct {
	contracts []extendedContract
}

// EncodeTo implements requestBody.
func (ecs extendedContractSet) EncodeTo(e *types.Encoder) {
	e.WriteUint64(uint64(len(ecs.contracts)))
	for _, ec := range ecs.contracts {
		ec.contract.Revision.EncodeTo(e)
		ec.contract.Signatures[0].EncodeTo(e)
		ec.contract.Signatures[1].EncodeTo(e)
		e.WriteUint64(ec.startHeight)
		ec.totalCost.EncodeTo(e)
		ec.uploadSpending.EncodeTo(e)
		ec.downloadSpending.EncodeTo(e)
		ec.fundAccountSpending.EncodeTo(e)
		ec.renewedFrom.EncodeTo(e)
	}
}

// DecodeFrom implements requestBody.
func (ecs extendedContractSet) DecodeFrom(d *types.Decoder) {
	// Nothing to do here.
}
