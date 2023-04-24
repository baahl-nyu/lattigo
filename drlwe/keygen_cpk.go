package drlwe

import (
	"io"

	"github.com/tuneinsight/lattigo/v4/ring"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/rlwe/ringqp"
	"github.com/tuneinsight/lattigo/v4/utils/sampling"
)

// CKGProtocol is the structure storing the parameters and and precomputations for the collective key generation protocol.
type CKGProtocol struct {
	params           rlwe.Parameters
	gaussianSamplerQ *ring.GaussianSampler
}

// ShallowCopy creates a shallow copy of CKGProtocol in which all the read-only data-structures are
// shared with the receiver and the temporary buffers are reallocated. The receiver and the returned
// CKGProtocol can be used concurrently.
func (ckg *CKGProtocol) ShallowCopy() *CKGProtocol {
	prng, err := sampling.NewPRNG()
	if err != nil {
		panic(err)
	}

	return &CKGProtocol{ckg.params, ring.NewGaussianSampler(prng, ckg.params.RingQ(), ckg.params.Sigma(), int(6*ckg.params.Sigma()))}
}

// CKGShare is a struct storing the CKG protocol's share.
type CKGShare struct {
	Value ringqp.Poly
}

// CKGCRP is a type for common reference polynomials in the CKG protocol.
type CKGCRP struct {
	Value ringqp.Poly
}

// BinarySize returns the size in bytes of the object
// when encoded using Encode.
func (share *CKGShare) BinarySize() int {
	return share.Value.BinarySize()
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (share *CKGShare) MarshalBinary() (p []byte, err error) {
	return share.Value.MarshalBinary()
}

// Encode encodes the object into a binary form on a preallocated slice of bytes
// and returns the number of bytes written.
func (share *CKGShare) Encode(p []byte) (ptr int, err error) {
	return share.Value.Encode(p)
}

// WriteTo writes the object on an io.Writer.
// To ensure optimal efficiency and minimal allocations, the user is encouraged
// to provide a struct implementing the interface buffer.Writer, which defines
// a subset of the method of the bufio.Writer.
// If w is not compliant to the buffer.Writer interface, it will be wrapped in
// a new bufio.Writer.
// For additional information, see lattigo/utils/buffer/writer.go.
func (share *CKGShare) WriteTo(w io.Writer) (n int64, err error) {
	return share.Value.WriteTo(w)
}

// UnmarshalBinary decodes a slice of bytes generated by
// MarshalBinary or WriteTo on the object.
func (share *CKGShare) UnmarshalBinary(p []byte) (err error) {
	return share.Value.UnmarshalBinary(p)
}

// Decode decodes a slice of bytes generated by Encode
// on the object and returns the number of bytes read.
func (share *CKGShare) Decode(p []byte) (n int, err error) {
	return share.Value.Decode(p)
}

// ReadFrom reads on the object from an io.Writer.
// To ensure optimal efficiency and minimal allocations, the user is encouraged
// to provide a struct implementing the interface buffer.Reader, which defines
// a subset of the method of the bufio.Reader.
// If r is not compliant to the buffer.Reader interface, it will be wrapped in
// a new bufio.Reader.
// For additional information, see lattigo/utils/buffer/reader.go.
func (share *CKGShare) ReadFrom(r io.Reader) (n int64, err error) {
	return share.Value.ReadFrom(r)
}

// NewCKGProtocol creates a new CKGProtocol instance
func NewCKGProtocol(params rlwe.Parameters) *CKGProtocol {
	ckg := new(CKGProtocol)
	ckg.params = params
	var err error
	prng, err := sampling.NewPRNG()
	if err != nil {
		panic(err)
	}
	ckg.gaussianSamplerQ = ring.NewGaussianSampler(prng, params.RingQ(), params.Sigma(), int(6*params.Sigma()))
	return ckg
}

// AllocateShare allocates the share of the CKG protocol.
func (ckg *CKGProtocol) AllocateShare() *CKGShare {
	return &CKGShare{*ckg.params.RingQP().NewPoly()}
}

// SampleCRP samples a common random polynomial to be used in the CKG protocol from the provided
// common reference string.
func (ckg *CKGProtocol) SampleCRP(crs CRS) CKGCRP {
	crp := ckg.params.RingQP().NewPoly()
	ringqp.NewUniformSampler(crs, *ckg.params.RingQP()).Read(crp)
	return CKGCRP{*crp}
}

// GenShare generates the party's public key share from its secret key as:
//
// crp*s_i + e_i
//
// for the receiver protocol. Has no effect is the share was already generated.
func (ckg *CKGProtocol) GenShare(sk *rlwe.SecretKey, crp CKGCRP, shareOut *CKGShare) {
	ringQP := ckg.params.RingQP()

	ckg.gaussianSamplerQ.Read(shareOut.Value.Q)

	if ringQP.RingP != nil {
		ringQP.ExtendBasisSmallNormAndCenter(shareOut.Value.Q, ckg.params.MaxLevelP(), nil, shareOut.Value.P)
	}

	ringQP.NTT(&shareOut.Value, &shareOut.Value)
	ringQP.MForm(&shareOut.Value, &shareOut.Value)

	ringQP.MulCoeffsMontgomeryThenSub(&sk.Value, &crp.Value, &shareOut.Value)
}

// AggregateShares aggregates a new share to the aggregate key
func (ckg *CKGProtocol) AggregateShares(share1, share2, shareOut *CKGShare) {
	ckg.params.RingQP().Add(&share1.Value, &share2.Value, &shareOut.Value)
}

// GenPublicKey return the current aggregation of the received shares as a bfv.PublicKey.
func (ckg *CKGProtocol) GenPublicKey(roundShare *CKGShare, crp CKGCRP, pubkey *rlwe.PublicKey) {
	pubkey.Value[0].Copy(&roundShare.Value)
	pubkey.Value[1].Copy(&crp.Value)
}
