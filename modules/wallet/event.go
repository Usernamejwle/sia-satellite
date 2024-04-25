package wallet

import (
	"fmt"
	"time"

	"github.com/mike76-dev/sia-satellite/modules"
	"go.sia.tech/core/consensus"
	"go.sia.tech/core/types"
)

// Event type constants.
const (
	EventTypeTransaction        = "transaction"
	EventTypeMinerPayout        = "miner payout"
	EventTypeMissedFileContract = "missed file contract"
)

// Annotate annotates a txpool transaction.
func Annotate(txn types.Transaction, ownsAddress func(types.Address) bool) modules.PoolTransaction {
	ptxn := modules.PoolTransaction{ID: txn.ID(), Raw: txn, Type: "unknown"}

	var totalValue types.Currency
	for _, sco := range txn.SiacoinOutputs {
		totalValue = totalValue.Add(sco.Value)
	}
	for _, fc := range txn.FileContracts {
		totalValue = totalValue.Add(fc.Payout)
	}
	for _, fee := range txn.MinerFees {
		totalValue = totalValue.Add(fee)
	}

	var ownedIn, ownedOut int
	for _, sci := range txn.SiacoinInputs {
		if ownsAddress(sci.UnlockConditions.UnlockHash()) {
			ownedIn++
		}
	}
	for _, sco := range txn.SiacoinOutputs {
		if ownsAddress(sco.Address) {
			ownedOut++
		}
	}
	var ins, outs string
	switch {
	case ownedIn == 0:
		ins = "none"
	case ownedIn < len(txn.SiacoinInputs):
		ins = "some"
	case ownedIn == len(txn.SiacoinInputs):
		ins = "all"
	}
	switch {
	case ownedOut == 0:
		outs = "none"
	case ownedOut < len(txn.SiacoinOutputs):
		outs = "some"
	case ownedOut == len(txn.SiacoinOutputs):
		outs = "all"
	}

	switch {
	case ins == "none" && outs == "none":
		ptxn.Type = "unrelated"
	case ins == "all":
		ptxn.Sent = totalValue
		switch {
		case outs == "all":
			ptxn.Type = "redistribution"
		case len(txn.FileContractRevisions) > 0:
			ptxn.Type = "contract revision"
		case len(txn.StorageProofs) > 0:
			ptxn.Type = "storage proof"
		case len(txn.ArbitraryData) > 0:
			ptxn.Type = "announcement"
		default:
			ptxn.Type = "send"
		}
	case ins == "none" && outs != "none":
		ptxn.Type = "receive"
		for _, sco := range txn.SiacoinOutputs {
			if ownsAddress(sco.Address) {
				ptxn.Received = ptxn.Received.Add(sco.Value)
			}
		}
	case ins == "some" && len(txn.FileContracts) > 0:
		ptxn.Type = "contract"
		for _, fc := range txn.FileContracts {
			var validLocked, missedLocked types.Currency
			for _, sco := range fc.ValidProofOutputs {
				if ownsAddress(sco.Address) {
					validLocked = validLocked.Add(fc.Payout)
				}
			}
			for _, sco := range fc.MissedProofOutputs {
				if ownsAddress(sco.Address) {
					missedLocked = missedLocked.Add(fc.Payout)
				}
			}
			if validLocked.Cmp(missedLocked) > 0 {
				ptxn.Locked = ptxn.Locked.Add(validLocked)
			} else {
				ptxn.Locked = ptxn.Locked.Add(missedLocked)
			}
		}
	}

	return ptxn
}

// An Event is something interesting that happened on the Sia blockchain.
type Event struct {
	Index     types.ChainIndex
	Timestamp time.Time
	Relevant  []types.Address
	Val       interface{ EventType() string }
}

// EventType implements Event.
func (*EventTransaction) EventType() string { return EventTypeTransaction }

// EventType implements Event.
func (*EventMinerPayout) EventType() string { return EventTypeMinerPayout }

// EventType implements Event.
func (*EventMissedFileContract) EventType() string { return EventTypeMissedFileContract }

// String implements fmt.Stringer.
func (e *Event) String() string {
	return fmt.Sprintf("%s at %s: %s", e.Val.EventType(), e.Timestamp, e.Val)
}

// A HostAnnouncement represents a host announcement within an EventTransaction.
type HostAnnouncement struct {
	PublicKey  types.PublicKey `json:"publicKey"`
	NetAddress string          `json:"netAddress"`
}

// A SiafundInput represents a siafund input within an EventTransaction.
type SiafundInput struct {
	SiafundElement types.SiafundElement `json:"siafundElement"`
	ClaimElement   types.SiacoinElement `json:"claimElement"`
}

// A FileContract represents a file contract within an EventTransaction.
type FileContract struct {
	FileContract types.FileContractElement `json:"fileContract"`
	// only non-nil if transaction revised contract
	Revision *types.FileContract `json:"revision,omitempty"`
	// only non-nil if transaction resolved contract
	ValidOutputs []types.SiacoinElement `json:"validOutputs,omitempty"`
}

// A V2FileContract represents a v2 file contract within an EventTransaction.
type V2FileContract struct {
	FileContract types.V2FileContractElement `json:"fileContract"`
	// only non-nil if transaction revised contract
	Revision *types.V2FileContract `json:"revision,omitempty"`
	// only non-nil if transaction resolved contract
	Resolution types.V2FileContractResolutionType `json:"resolution,omitempty"`
	Outputs    []types.SiacoinElement             `json:"outputs,omitempty"`
}

// An EventTransaction represents a transaction that affects the wallet.
type EventTransaction struct {
	ID                types.TransactionID    `json:"id"`
	SiacoinInputs     []types.SiacoinElement `json:"siacoinInputs"`
	SiacoinOutputs    []types.SiacoinElement `json:"siacoinOutputs"`
	SiafundInputs     []SiafundInput         `json:"siafundInputs"`
	SiafundOutputs    []types.SiafundElement `json:"siafundOutputs"`
	FileContracts     []FileContract         `json:"fileContracts"`
	V2FileContracts   []V2FileContract       `json:"v2FileContracts"`
	HostAnnouncements []HostAnnouncement     `json:"hostAnnouncements"`
	Fee               types.Currency         `json:"fee"`
}

// An EventMinerPayout represents a miner payout from a block.
type EventMinerPayout struct {
	SiacoinOutput types.SiacoinElement `json:"siacoinOutput"`
}

// An EventMissedFileContract represents a file contract that has expired
// without a storage proof
type EventMissedFileContract struct {
	FileContract  types.FileContractElement `json:"fileContract"`
	MissedOutputs []types.SiacoinElement    `json:"missedOutputs"`
}

// String implements fmt.Stringer.
func (et *EventTransaction) String() string {
	result := et.ID.String()
	if len(et.SiacoinOutputs) > 0 {
		result += ": Siacoin outputs: "
	}
	for i, sco := range et.SiacoinOutputs {
		result += sco.SiacoinOutput.Address.String()
		result += fmt.Sprintf(" (%s)", sco.SiacoinOutput.Value)
		if i < len(et.SiacoinOutputs)-1 {
			result += ", "
		}
	}
	if len(et.SiafundOutputs) > 0 {
		result += "; Siafund outputs: "
	}
	for i, sfo := range et.SiafundOutputs {
		result += sfo.SiafundOutput.Address.String()
		result += fmt.Sprintf(" (%d SF)", sfo.SiafundOutput.Value)
		if i < len(et.SiafundOutputs)-1 {
			result += ", "
		}
	}
	return result
}

// String implements fmt.Stringer.
func (emp *EventMinerPayout) String() string {
	return fmt.Sprintf("%s (%s)",
		emp.SiacoinOutput.SiacoinOutput.Address.String(),
		emp.SiacoinOutput.SiacoinOutput.Value,
	)
}

// String implements fmt.Stringer.
func (emfc *EventMissedFileContract) String() string {
	return emfc.FileContract.ID.String()
}

// A ChainUpdate is a set of changes to the consensus state.
type ChainUpdate interface {
	ForEachSiacoinElement(func(sce types.SiacoinElement, spent bool))
	ForEachSiafundElement(func(sfe types.SiafundElement, spent bool))
	ForEachFileContractElement(func(fce types.FileContractElement, rev *types.FileContractElement, resolved, valid bool))
	ForEachV2FileContractElement(func(fce types.V2FileContractElement, rev *types.V2FileContractElement, res types.V2FileContractResolutionType))
}

// AppliedEvents extracts a list of relevant events from a chain update.
func AppliedEvents(cs consensus.State, b types.Block, cu ChainUpdate, relevant func(types.Address) bool) []Event {
	var events []Event
	addEvent := func(v interface{ EventType() string }, relevant []types.Address) {
		// Dedup relevant addresses.
		seen := make(map[types.Address]bool)
		unique := relevant[:0]
		for _, addr := range relevant {
			if !seen[addr] {
				unique = append(unique, addr)
				seen[addr] = true
			}
		}

		events = append(events, Event{
			Timestamp: b.Timestamp,
			Index:     cs.Index,
			Relevant:  unique,
			Val:       v,
		})
	}

	// Do a first pass to see if there's anything relevant in the block.
	relevantContract := func(fc types.FileContract) (addrs []types.Address) {
		for _, sco := range fc.ValidProofOutputs {
			if relevant(sco.Address) {
				addrs = append(addrs, sco.Address)
			}
		}
		for _, sco := range fc.MissedProofOutputs {
			if relevant(sco.Address) {
				addrs = append(addrs, sco.Address)
			}
		}
		return
	}
	relevantV2Contract := func(fc types.V2FileContract) (addrs []types.Address) {
		if relevant(fc.RenterOutput.Address) {
			addrs = append(addrs, fc.RenterOutput.Address)
		}
		if relevant(fc.HostOutput.Address) {
			addrs = append(addrs, fc.HostOutput.Address)
		}
		return
	}
	relevantV2ContractResolution := func(res types.V2FileContractResolutionType) (addrs []types.Address) {
		switch r := res.(type) {
		case *types.V2FileContractFinalization:
			return relevantV2Contract(types.V2FileContract(*r))
		case *types.V2FileContractRenewal:
			return relevantV2Contract(r.FinalRevision)
		}
		return
	}
	anythingRelevant := func() (ok bool) {
		cu.ForEachSiacoinElement(func(sce types.SiacoinElement, spent bool) {
			if ok || relevant(sce.SiacoinOutput.Address) {
				ok = true
			}
		})
		cu.ForEachSiafundElement(func(sfe types.SiafundElement, spent bool) {
			if ok || relevant(sfe.SiafundOutput.Address) {
				ok = true
			}
		})
		cu.ForEachFileContractElement(func(fce types.FileContractElement, rev *types.FileContractElement, resolved, valid bool) {
			if ok || len(relevantContract(fce.FileContract)) > 0 || (rev != nil && len(relevantContract(rev.FileContract)) > 0) {
				ok = true
			}
		})
		cu.ForEachV2FileContractElement(func(fce types.V2FileContractElement, rev *types.V2FileContractElement, res types.V2FileContractResolutionType) {
			if ok ||
				len(relevantV2Contract(fce.V2FileContract)) > 0 ||
				(rev != nil && len(relevantV2Contract(rev.V2FileContract)) > 0) ||
				(res != nil && len(relevantV2ContractResolution(res)) > 0) {
				ok = true
			}
		})
		return
	}()
	if !anythingRelevant {
		return nil
	}

	// Collect all elements.
	sces := make(map[types.SiacoinOutputID]types.SiacoinElement)
	sfes := make(map[types.SiafundOutputID]types.SiafundElement)
	fces := make(map[types.FileContractID]types.FileContractElement)
	v2fces := make(map[types.FileContractID]types.V2FileContractElement)
	cu.ForEachSiacoinElement(func(sce types.SiacoinElement, spent bool) {
		sce.MerkleProof = nil
		sces[types.SiacoinOutputID(sce.ID)] = sce
	})
	cu.ForEachSiafundElement(func(sfe types.SiafundElement, spent bool) {
		sfe.MerkleProof = nil
		sfes[types.SiafundOutputID(sfe.ID)] = sfe
	})
	cu.ForEachFileContractElement(func(fce types.FileContractElement, rev *types.FileContractElement, resolved, valid bool) {
		fce.MerkleProof = nil
		fces[types.FileContractID(fce.ID)] = fce
	})
	cu.ForEachV2FileContractElement(func(fce types.V2FileContractElement, rev *types.V2FileContractElement, res types.V2FileContractResolutionType) {
		fce.MerkleProof = nil
		v2fces[types.FileContractID(fce.ID)] = fce
	})

	relevantTxn := func(txn types.Transaction) (addrs []types.Address) {
		for _, sci := range txn.SiacoinInputs {
			if sce := sces[sci.ParentID]; relevant(sce.SiacoinOutput.Address) {
				addrs = append(addrs, sce.SiacoinOutput.Address)
			}
		}
		for _, sco := range txn.SiacoinOutputs {
			if relevant(sco.Address) {
				addrs = append(addrs, sco.Address)
			}
		}
		for _, sfi := range txn.SiafundInputs {
			if sfe := sfes[sfi.ParentID]; relevant(sfe.SiafundOutput.Address) {
				addrs = append(addrs, sfe.SiafundOutput.Address)
			}
		}
		for _, sfo := range txn.SiafundOutputs {
			if relevant(sfo.Address) {
				addrs = append(addrs, sfo.Address)
			}
		}
		for _, fc := range txn.FileContracts {
			addrs = append(addrs, relevantContract(fc)...)
		}
		for _, fcr := range txn.FileContractRevisions {
			addrs = append(addrs, relevantContract(fcr.FileContract)...)
		}
		for _, sp := range txn.StorageProofs {
			addrs = append(addrs, relevantContract(fces[sp.ParentID].FileContract)...)
		}
		return
	}

	relevantV2Txn := func(txn types.V2Transaction) (addrs []types.Address) {
		for _, sci := range txn.SiacoinInputs {
			if relevant(sci.Parent.SiacoinOutput.Address) {
				addrs = append(addrs, sci.Parent.SiacoinOutput.Address)
			}
		}
		for _, sco := range txn.SiacoinOutputs {
			if relevant(sco.Address) {
				addrs = append(addrs, sco.Address)
			}
		}
		for _, sfi := range txn.SiafundInputs {
			if relevant(sfi.Parent.SiafundOutput.Address) {
				addrs = append(addrs, sfi.Parent.SiafundOutput.Address)
			}
		}
		for _, sfo := range txn.SiafundOutputs {
			if relevant(sfo.Address) {
				addrs = append(addrs, sfo.Address)
			}
		}
		for _, fc := range txn.FileContracts {
			addrs = append(addrs, relevantV2Contract(fc)...)
		}
		for _, fcr := range txn.FileContractRevisions {
			addrs = append(addrs, relevantV2Contract(fcr.Parent.V2FileContract)...)
			addrs = append(addrs, relevantV2Contract(fcr.Revision)...)
		}
		for _, fcr := range txn.FileContractResolutions {
			addrs = append(addrs, relevantV2Contract(fcr.Parent.V2FileContract)...)
			switch r := fcr.Resolution.(type) {
			case *types.V2FileContractFinalization:
				addrs = append(addrs, relevantV2Contract(types.V2FileContract(*r))...)
			case *types.V2FileContractRenewal:
				addrs = append(addrs, relevantV2Contract(r.FinalRevision)...)
			}
		}
		return
	}

	// Handle v1 transactions.
	for _, txn := range b.Transactions {
		relevant := relevantTxn(txn)
		if len(relevant) == 0 {
			continue
		}

		e := &EventTransaction{
			ID:             txn.ID(),
			SiacoinInputs:  make([]types.SiacoinElement, len(txn.SiacoinInputs)),
			SiacoinOutputs: make([]types.SiacoinElement, len(txn.SiacoinOutputs)),
			SiafundInputs:  make([]SiafundInput, len(txn.SiafundInputs)),
			SiafundOutputs: make([]types.SiafundElement, len(txn.SiafundOutputs)),
		}

		for i := range txn.SiacoinInputs {
			e.SiacoinInputs[i] = sces[txn.SiacoinInputs[i].ParentID]
		}
		for i := range txn.SiacoinOutputs {
			e.SiacoinOutputs[i] = sces[txn.SiacoinOutputID(i)]
		}
		for i := range txn.SiafundInputs {
			e.SiafundInputs[i] = SiafundInput{
				SiafundElement: sfes[txn.SiafundInputs[i].ParentID],
				ClaimElement:   sces[txn.SiafundClaimOutputID(i)],
			}
		}
		for i := range txn.SiafundOutputs {
			e.SiafundOutputs[i] = sfes[txn.SiafundOutputID(i)]
		}
		addContract := func(id types.FileContractID) *FileContract {
			for i := range e.FileContracts {
				if types.FileContractID(e.FileContracts[i].FileContract.ID) == id {
					return &e.FileContracts[i]
				}
			}
			e.FileContracts = append(e.FileContracts, FileContract{FileContract: fces[id]})
			return &e.FileContracts[len(e.FileContracts)-1]
		}
		for i := range txn.FileContracts {
			addContract(txn.FileContractID(i))
		}
		for i := range txn.FileContractRevisions {
			fc := addContract(txn.FileContractRevisions[i].ParentID)
			rev := txn.FileContractRevisions[i].FileContract
			fc.Revision = &rev
		}
		for i := range txn.StorageProofs {
			fc := addContract(txn.StorageProofs[i].ParentID)
			fc.ValidOutputs = make([]types.SiacoinElement, len(fc.FileContract.FileContract.ValidProofOutputs))
			for i := range fc.ValidOutputs {
				fc.ValidOutputs[i] = sces[types.FileContractID(fc.FileContract.ID).ValidOutputID(i)]
			}
		}
		for _, arb := range txn.ArbitraryData {
			var prefix types.Specifier
			var uk types.UnlockKey
			d := types.NewBufDecoder(arb)
			prefix.DecodeFrom(d)
			netAddress := d.ReadString()
			uk.DecodeFrom(d)
			if d.Err() == nil && prefix == types.NewSpecifier("HostAnnouncement") &&
				uk.Algorithm == types.SpecifierEd25519 && len(uk.Key) == len(types.PublicKey{}) {
				e.HostAnnouncements = append(e.HostAnnouncements, HostAnnouncement{
					PublicKey:  *(*types.PublicKey)(uk.Key),
					NetAddress: netAddress,
				})
			}
		}
		for i := range txn.MinerFees {
			e.Fee = e.Fee.Add(txn.MinerFees[i])
		}

		addEvent(e, relevant)
	}

	// Handle v2 transactions.
	for _, txn := range b.V2Transactions() {
		relevant := relevantV2Txn(txn)
		if len(relevant) == 0 {
			continue
		}

		txid := txn.ID()
		e := &EventTransaction{
			ID:             txid,
			SiacoinInputs:  make([]types.SiacoinElement, len(txn.SiacoinInputs)),
			SiacoinOutputs: make([]types.SiacoinElement, len(txn.SiacoinOutputs)),
			SiafundInputs:  make([]SiafundInput, len(txn.SiafundInputs)),
			SiafundOutputs: make([]types.SiafundElement, len(txn.SiafundOutputs)),
		}
		for i := range txn.SiacoinInputs {
			// NOTE: here (and elsewhere), we fetch the element from our maps,
			// rather than using the parent directly, because our copy has its
			// Merkle proof nil'd out.
			e.SiacoinInputs[i] = sces[types.SiacoinOutputID(txn.SiacoinInputs[i].Parent.ID)]
		}
		for i := range txn.SiacoinOutputs {
			e.SiacoinOutputs[i] = sces[txn.SiacoinOutputID(txid, i)]
		}
		for i := range txn.SiafundInputs {
			sfoid := types.SiafundOutputID(txn.SiafundInputs[i].Parent.ID)
			e.SiafundInputs[i] = SiafundInput{
				SiafundElement: sfes[sfoid],
				ClaimElement:   sces[sfoid.ClaimOutputID()],
			}
		}
		for i := range txn.SiafundOutputs {
			e.SiafundOutputs[i] = sfes[txn.SiafundOutputID(txid, i)]
		}
		addContract := func(id types.FileContractID) *V2FileContract {
			for i := range e.V2FileContracts {
				if types.FileContractID(e.V2FileContracts[i].FileContract.ID) == id {
					return &e.V2FileContracts[i]
				}
			}
			e.V2FileContracts = append(e.V2FileContracts, V2FileContract{FileContract: v2fces[id]})
			return &e.V2FileContracts[len(e.V2FileContracts)-1]
		}
		for i := range txn.FileContracts {
			addContract(txn.V2FileContractID(txid, i))
		}
		for _, fcr := range txn.FileContractRevisions {
			fc := addContract(types.FileContractID(fcr.Parent.ID))
			fc.Revision = &fcr.Revision
		}
		for _, fcr := range txn.FileContractResolutions {
			fc := addContract(types.FileContractID(fcr.Parent.ID))
			fc.Resolution = fcr.Resolution
			fc.Outputs = []types.SiacoinElement{
				sces[types.FileContractID(fcr.Parent.ID).V2RenterOutputID()],
				sces[types.FileContractID(fcr.Parent.ID).V2HostOutputID()],
			}
		}
		for _, a := range txn.Attestations {
			if a.Key == "HostAnnouncement" {
				e.HostAnnouncements = append(e.HostAnnouncements, HostAnnouncement{
					PublicKey:  a.PublicKey,
					NetAddress: string(a.Value),
				})
			}
		}

		e.Fee = txn.MinerFee
		addEvent(e, relevant)
	}

	// Handle missed contracts.
	cu.ForEachFileContractElement(func(fce types.FileContractElement, rev *types.FileContractElement, resolved, valid bool) {
		if resolved && !valid {
			relevant := relevantContract(fce.FileContract)
			if len(relevant) == 0 {
				return
			}
			missedOutputs := make([]types.SiacoinElement, len(fce.FileContract.MissedProofOutputs))
			for i := range missedOutputs {
				missedOutputs[i] = sces[types.FileContractID(fce.ID).MissedOutputID(i)]
			}
			addEvent(&EventMissedFileContract{
				FileContract:  fce,
				MissedOutputs: missedOutputs,
			}, relevant)
		}
	})

	// Handle block rewards.
	for i := range b.MinerPayouts {
		if relevant(b.MinerPayouts[i].Address) {
			addEvent(&EventMinerPayout{
				SiacoinOutput: sces[cs.Index.ID.MinerOutputID(i)],
			}, []types.Address{b.MinerPayouts[i].Address})
		}
	}

	return events
}
