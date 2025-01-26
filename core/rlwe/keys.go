package rlwe

import (
	"bufio"
	"fmt"
	"io"
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/baahl-nyu/lattigo/v6/ring/ringqp"
	"github.com/baahl-nyu/lattigo/v6/utils/buffer"
	"github.com/baahl-nyu/lattigo/v6/utils/sampling"
	"github.com/baahl-nyu/lattigo/v6/utils/structs"
)

// SecretKey is a type for generic RLWE secret keys.
// The Value field stores the polynomial in NTT and Montgomery form.
type SecretKey struct {
	Value ringqp.Poly
}

// NewSecretKey generates a new [SecretKey] with zero values.
func NewSecretKey(params ParameterProvider) *SecretKey {
	return &SecretKey{Value: params.GetRLWEParameters().RingQP().NewPoly()}
}

func (sk SecretKey) Equal(other *SecretKey) bool {
	return cmp.Equal(sk.Value, other.Value)
}

// LevelQ returns the level of the modulus Q of the target.
func (sk SecretKey) LevelQ() int {
	return sk.Value.Q.Level()
}

// LevelP returns the level of the modulus P of the target.
// Returns -1 if P is absent.
func (sk SecretKey) LevelP() int {
	return sk.Value.P.Level()
}

// CopyNew creates a deep copy of the receiver [SecretKey] and returns it.
func (sk SecretKey) CopyNew() *SecretKey {
	return &SecretKey{*sk.Value.CopyNew()}
}

// BinarySize returns the serialized size of the object in bytes.
func (sk SecretKey) BinarySize() (dataLen int) {
	return sk.Value.BinarySize()
}

// WriteTo writes the object on an [io.Writer]. It implements the [io.WriterTo]
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the [buffer.Writer] interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a [bufio.Writer]. Since this requires allocations, it
// is preferable to pass a [buffer.Writer] directly:
//
//   - When writing multiple times to a [io.Writer], it is preferable to first wrap the
//     io.Writer in a pre-allocated [bufio.Writer].
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (sk SecretKey) WriteTo(w io.Writer) (n int64, err error) {
	return sk.Value.WriteTo(w)
}

// ReadFrom reads on the object from an [io.Writer]. It implements the
// [io.ReaderFrom] interface.
//
// Unless r implements the [buffer.Reader] interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a [bufio.Reader]. Since this requires allocation, it
// is preferable to pass a [buffer.Reader] directly:
//
//   - When reading multiple values from a [io.Reader], it is preferable to first
//     first wrap io.Reader in a pre-allocated [bufio.Reader].
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (sk *SecretKey) ReadFrom(r io.Reader) (n int64, err error) {
	return sk.Value.ReadFrom(r)
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (sk SecretKey) MarshalBinary() (p []byte, err error) {
	return sk.Value.MarshalBinary()
}

// UnmarshalBinary decodes a slice of bytes generated by
// [SecretKey.MarshalBinary] or [SecretKey.WriteTo] on the object.
func (sk *SecretKey) UnmarshalBinary(p []byte) (err error) {
	return sk.Value.UnmarshalBinary(p)
}

func (sk *SecretKey) isEncryptionKey() {}

type VectorQP []ringqp.Poly

// NewVectorQP returns a new [PublicKey] with zero values.
func NewVectorQP(params ParameterProvider, size, levelQ, levelP int) (v VectorQP) {
	rqp := params.GetRLWEParameters().RingQP().AtLevel(levelQ, levelP)

	v = make(VectorQP, size)

	for i := range v {
		v[i] = rqp.NewPoly()
	}

	return
}

// LevelQ returns the level of the modulus Q of the first element of the [VectorQP].
// Returns -1 if the size of the vector is zero or has no modulus Q.
func (p VectorQP) LevelQ() int {
	if len(p) == 0 {
		return -1
	}
	return p[0].LevelQ()
}

// LevelP returns the level of the modulus P of the first element of the [VectorQP].
// Returns -1 if the size of the vector is zero or has no modulus P.
func (p VectorQP) LevelP() int {
	if len(p) == 0 {
		return -1
	}
	return p[0].LevelP()
}

// CopyNew creates a deep copy of the target [PublicKey] and returns it.
func (p VectorQP) CopyNew() *VectorQP {
	m := make([]ringqp.Poly, len(p))
	for i := range p {
		m[i] = *p[i].CopyNew()
	}
	v := VectorQP(m)
	return &v
}

// Equal performs a deep equal.
func (p VectorQP) Equal(other *VectorQP) (equal bool) {

	if len(p) != len(*other) {
		return false
	}

	equal = true
	for i := range p {
		equal = equal && p[i].Equal(&(*other)[i])
	}

	return
}

func (p VectorQP) BinarySize() int {
	return structs.Vector[ringqp.Poly](p[:]).BinarySize()
}

// WriteTo writes the object on an [io.Writer]. It implements the [io.WriterTo]
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the [buffer.Writer] interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a [bufio.Writer]. Since this requires allocations, it
// is preferable to pass a [buffer.Writer] directly:
//
//   - When writing multiple times to a [io.Writer], it is preferable to first wrap the
//     io.Writer in a pre-allocated [bufio.Writer].
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (p VectorQP) WriteTo(w io.Writer) (n int64, err error) {
	v := structs.Vector[ringqp.Poly](p[:])
	return v.WriteTo(w)
}

// ReadFrom reads on the object from an [io.Writer]. It implements the
// [io.ReaderFrom] interface.
//
// Unless r implements the [buffer.Reader] interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a [bufio.Reader]. Since this requires allocation, it
// is preferable to pass a [buffer.Reader] directly:
//
//   - When reading multiple values from a [io.Reader], it is preferable to first
//     first wrap io.Reader in a pre-allocated [bufio.Reader].
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (p *VectorQP) ReadFrom(r io.Reader) (n int64, err error) {
	v := structs.Vector[ringqp.Poly](*p)
	n, err = v.ReadFrom(r)
	*p = VectorQP(v)
	return
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (p VectorQP) MarshalBinary() ([]byte, error) {
	buf := buffer.NewBufferSize(p.BinarySize())
	_, err := p.WriteTo(buf)
	return buf.Bytes(), err
}

// UnmarshalBinary decodes a slice of bytes generated by
// [VectorQP.MarshalBinary] or [VectorQP.WriteTo] on the object.
func (p *VectorQP) UnmarshalBinary(b []byte) error {
	_, err := p.ReadFrom(buffer.NewBuffer(b))
	return err
}

// PublicKey is a type for generic RLWE public keys.
// The Value field stores the polynomials in NTT and Montgomery form.
type PublicKey struct {
	Value VectorQP
}

// NewPublicKey returns a new [PublicKey] with zero values.
func NewPublicKey(params ParameterProvider) (pk *PublicKey) {
	p := params.GetRLWEParameters()
	return &PublicKey{Value: NewVectorQP(params, 2, p.MaxLevelQ(), p.MaxLevelP())}
}

func (p PublicKey) LevelQ() int {
	return p.Value.LevelQ()
}

func (p PublicKey) LevelP() int {
	return p.Value.LevelP()
}

// CopyNew creates a deep copy of the target [PublicKey] and returns it.
func (p PublicKey) CopyNew() *PublicKey {
	return &PublicKey{Value: *p.Value.CopyNew()}
}

// Equal performs a deep equal.
func (p PublicKey) Equal(other *PublicKey) bool {
	return p.Value.Equal(&other.Value)
}

func (p PublicKey) BinarySize() int {
	return p.Value.BinarySize()
}

// WriteTo writes the object on an [io.Writer]. It implements the [io.WriterTo]
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the [buffer.Writer] interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a [bufio.Writer]. Since this requires allocations, it
// is preferable to pass a [buffer.Writer] directly:
//
//   - When writing multiple times to a [io.Writer], it is preferable to first wrap the
//     io.Writer in a pre-allocated [bufio.Writer].
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (p PublicKey) WriteTo(w io.Writer) (n int64, err error) {
	return p.Value.WriteTo(w)
}

// ReadFrom reads on the object from an [io.Writer]. It implements the
// [io.ReaderFrom] interface.
//
// Unless r implements the [buffer.Reader] interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a [bufio.Reader]. Since this requires allocation, it
// is preferable to pass a [buffer.Reader] directly:
//
//   - When reading multiple values from a [io.Reader], it is preferable to first
//     first wrap io.Reader in a pre-allocated [bufio.Reader].
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (p *PublicKey) ReadFrom(r io.Reader) (n int64, err error) {
	return p.Value.ReadFrom(r)
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (p PublicKey) MarshalBinary() ([]byte, error) {
	return p.Value.MarshalBinary()
}

// UnmarshalBinary decodes a slice of bytes generated by
// [PublicKey.MarshalBinary] or [PublicKey.WriteTo] on the object.
func (p *PublicKey) UnmarshalBinary(b []byte) error {
	return p.Value.UnmarshalBinary(b)
}

func (p *PublicKey) isEncryptionKey() {}

// EvaluationKey is a public key indented to be used during the evaluation phase of a homomorphic circuit.
// It provides a one way public and non-interactive re-encryption from a ciphertext encrypted under skIn
// to a ciphertext encrypted under skOut.
//
// Such re-encryption is for example used for:
//   - Homomorphic relinearization: re-encryption of a quadratic ciphertext (that requires (1, sk sk^2) to be decrypted)
//     to a linear ciphertext (that required (1, sk) to be decrypted). In this case skIn = sk^2 an skOut = sk.
//   - Homomorphic automorphisms: an automorphism in the ring Z[X]/(X^{N}+1) is defined as pi_k: X^{i} -> X^{i^k} with
//     k coprime to 2N. Pi_sk is for exampled used during homomorphic slot rotations. Applying pi_k to a ciphertext encrypted
//     under sk generates a new ciphertext encrypted under pi_k(sk), and an Evaluationkey skIn = pi_k(sk) to skOut = sk
//     is used to bring it back to its original key.
type EvaluationKey struct {
	GadgetCiphertext
	Seed *[32]byte // Must be != nil iff EvaluationKey.IsCompressed() = true
}

type EvaluationKeyParameters struct {
	LevelQ               *int
	LevelP               *int
	BaseTwoDecomposition *int
	Compressed           bool
}

func ResolveEvaluationKeyParameters(params Parameters, evkParams []EvaluationKeyParameters) (levelQ, levelP, BaseTwoDecomposition int, compressed bool) {
	if len(evkParams) != 0 {
		if evkParams[0].LevelQ == nil {
			levelQ = params.MaxLevelQ()
		} else {
			levelQ = *evkParams[0].LevelQ
		}

		if evkParams[0].LevelP == nil {
			levelP = params.MaxLevelP()
		} else {
			levelP = *evkParams[0].LevelP
		}

		if evkParams[0].BaseTwoDecomposition != nil {
			BaseTwoDecomposition = *evkParams[0].BaseTwoDecomposition
		}
		compressed = evkParams[0].Compressed
	} else {
		levelQ = params.MaxLevelQ()
		levelP = params.MaxLevelP()
	}

	return
}

// NewEvaluationKey returns a new [EvaluationKey] with pre-allocated zero-value.
func NewEvaluationKey(params ParameterProvider, evkParams ...EvaluationKeyParameters) *EvaluationKey {
	p := *params.GetRLWEParameters()
	levelQ, levelP, BaseTwoDecomposition, compressed := ResolveEvaluationKeyParameters(p, evkParams)
	return newEvaluationKey(p, levelQ, levelP, BaseTwoDecomposition, compressed)
}

func newEvaluationKey(params Parameters, levelQ, levelP, BaseTwoDecomposition int, compressed bool) *EvaluationKey {
	degree := 1
	if compressed {
		degree = 0
	}
	return &EvaluationKey{
		GadgetCiphertext: *NewGadgetCiphertext(params, degree, levelQ, levelP, BaseTwoDecomposition),
	}
}

// IsCompressed indicates whether the [EvaluationKey] is compressed or not.
func (evk EvaluationKey) IsCompressed() bool {
	return evk.Degree() == 0
}

// Expand decompresses a compressed [EvaluationKey] of the form (-a*sk + w*P*s' + e) to (-a*sk + w*P*s' + e, a).
// The user can provide a buffer GadgetCiphertext of degree 0 matching the size of the [EvaluationKey].
// If no buffer is provided, the second component will be allocated.
// The method will return an error if:
//   - The [EvaluationKey] is not compressed
//   - The provided buffer is invalid.
func (evk EvaluationKey) Expand(params ParameterProvider, buffer *GadgetCiphertext) error {

	if !evk.IsCompressed() {
		return fmt.Errorf("evaluation key is not compressed")
	}

	if evk.Seed == nil {
		return fmt.Errorf("seed is missing")
	}

	prng, err := sampling.NewKeyedPRNG((*evk.Seed)[:])
	if err != nil {
		panic(fmt.Errorf("sampling.NewKeyedPRNG: %s", err))
	}

	levelQ := evk.LevelQ()
	levelP := evk.LevelP()

	uniformRingQPSampler := ringqp.NewUniformSampler(prng, *params.GetRLWEParameters().RingQP()).AtLevel(levelQ, levelP)
	BaseRNSDecompositionVectorSize := evk.BaseRNSDecompositionVectorSize()
	BaseTwoDecompositionVectorSize := evk.BaseTwoDecompositionVectorSize()

	// Check that the provided buffer is of the correct size
	if buffer != nil {

		if have := buffer.Degree(); have != 0 {
			return fmt.Errorf("invalid buffer, degree should be 0 but is %d", have)
		}

		if have := buffer.BaseRNSDecompositionVectorSize(); have != BaseRNSDecompositionVectorSize {
			return fmt.Errorf("invalid buffer BaseRNSDecompositionVectorSize, should be %d but is %d", have, BaseRNSDecompositionVectorSize)
		}

		if have := buffer.BaseTwoDecompositionVectorSize(); !slices.Equal(have, BaseTwoDecompositionVectorSize) {
			return fmt.Errorf("invalid buffer BaseTwoDecompositionVectorSize, should be %v but is %v", have, BaseTwoDecompositionVectorSize)
		}

		if have := buffer.LevelQ(); have != levelQ {
			return fmt.Errorf("invalid buffer levelQ, should be %d but is %d", levelQ, have)
		}

		if have := buffer.LevelP(); have != levelP {
			return fmt.Errorf("invalid buffer levelP, should be %d but is %d", levelP, have)
		}

	} else {
		buffer = NewGadgetCiphertext(params, 0, levelQ, levelP, evk.BaseTwoDecomposition)
	}

	// This works because the uniform RingQP sampler is only used to sample 'a'
	// during the creation of the compressed evaluation key with no other call to
	// the sampler. Hence, both PRNG invocation sequences are equal.
	value := make(structs.Matrix[VectorQP], BaseRNSDecompositionVectorSize)
	for i := 0; i < BaseRNSDecompositionVectorSize; i++ {
		value[i] = make([]VectorQP, BaseTwoDecompositionVectorSize[i])
		for j := range value[i] {
			// Sample 'a' and create the full evaluation key (-a*sk + w*P*s' + e, a).
			uniformRingQPSampler.Read(buffer.Value[i][j][0])
			evk.Value[i][j] = VectorQP{evk.Value[i][j][0], buffer.Value[i][j][0]}
		}
	}

	return nil
}

// BinarySize returns the serialized size of the object in bytes.
func (evk EvaluationKey) BinarySize() (size int) {
	if evk.Seed != nil {
		return evk.GadgetCiphertext.BinarySize() + len(*evk.Seed)
	}
	return evk.GadgetCiphertext.BinarySize()
}

// WriteTo writes the object on an [io.Writer]. It implements the [io.WriterTo]
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the [buffer.Writer] interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a [bufio.Writer]. Since this requires allocations, it
// is preferable to pass a [buffer.Writer] directly:
//
//   - When writing multiple times to a [io.Writer], it is preferable to first wrap the
//     io.Writer in a pre-allocated [bufio.Writer].
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (evk EvaluationKey) WriteTo(w io.Writer) (n int64, err error) {
	switch w := w.(type) {
	case buffer.Writer:

		var inc int64

		if inc, err = evk.GadgetCiphertext.WriteTo(w); err != nil {
			return n + inc, err
		}

		n += inc

		if evk.IsCompressed() {

			// Sanity check, should not happen unless evk has been manually modified
			if evk.Seed == nil {
				return n + inc, fmt.Errorf("writing compressed evaluation key: the seed is nil")
			}

			if inc, err = buffer.Write(w, (*evk.Seed)[:]); err != nil {
				return n + inc, err
			}

			n += inc
		}

		if err = w.Flush(); err != nil {
			return n, err
		}
		return

	default:
		return evk.WriteTo(bufio.NewWriter(w))
	}
}

// ReadFrom reads on the object from an [io.Writer]. It implements the
// [io.ReaderFrom] interface.
//
// Unless r implements the [buffer.Reader] interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a [bufio.Reader]. Since this requires allocation, it
// is preferable to pass a [buffer.Reader] directly:
//
//   - When reading multiple values from a [io.Reader], it is preferable to first
//     first wrap io.Reader in a pre-allocated [bufio.Reader].
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (evk *EvaluationKey) ReadFrom(r io.Reader) (n int64, err error) {
	switch r := r.(type) {
	case buffer.Reader:

		var inc int64

		if inc, err = evk.GadgetCiphertext.ReadFrom(r); err != nil {
			return n + inc, err
		}

		n += inc

		if evk.IsCompressed() {
			var seed [32]byte
			if inc, err = buffer.Read(r, seed[:]); err != nil {
				return n + inc, err
			}

			evk.Seed = &seed

			n += inc
		}

		return
	default:
		return evk.ReadFrom(bufio.NewReader(r))
	}
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (evk EvaluationKey) MarshalBinary() (p []byte, err error) {
	buf := buffer.NewBufferSize(evk.BinarySize())
	_, err = evk.WriteTo(buf)
	return buf.Bytes(), err
}

// UnmarshalBinary decodes a slice of bytes generated by
// [EvaluationKey.MarshalBinary] or [EvaluationKey.WriteTo] on the object.
func (evk *EvaluationKey) UnmarshalBinary(p []byte) (err error) {
	_, err = evk.ReadFrom(buffer.NewBuffer(p))
	return
}

// CopyNew creates a deep copy of the target [EvaluationKey] and returns it.
func (evk EvaluationKey) CopyNew() *EvaluationKey {
	return &EvaluationKey{GadgetCiphertext: *evk.GadgetCiphertext.CopyNew()}
}

// RelinearizationKey is type of [EvaluationKey] used for ciphertext multiplication compactness.
// The Relinearization key encrypts s^{2} under s and is used to homomorphically re-encrypt the
// degree 2 term of a ciphertext (the term that decrypt with s^{2}) into a degree 1 term
// (a term that decrypts with s).
type RelinearizationKey struct {
	EvaluationKey
}

// NewRelinearizationKey allocates a new [RelinearizationKey] with zero coefficients.
func NewRelinearizationKey(params ParameterProvider, evkParams ...EvaluationKeyParameters) *RelinearizationKey {
	p := *params.GetRLWEParameters()
	levelQ, levelP, BaseTwoDecomposition, compressed := ResolveEvaluationKeyParameters(p, evkParams)
	return newRelinearizationKey(p, levelQ, levelP, BaseTwoDecomposition, compressed)
}

func newRelinearizationKey(params Parameters, levelQ, levelP, BaseTwoDecomposition int, compressed bool) *RelinearizationKey {
	degree := 1
	if compressed {
		degree = 0
	}
	return &RelinearizationKey{EvaluationKey: EvaluationKey{GadgetCiphertext: *NewGadgetCiphertext(params, degree, levelQ, levelP, BaseTwoDecomposition)}}
}

// CopyNew creates a deep copy of the object and returns it.
func (rlk RelinearizationKey) CopyNew() *RelinearizationKey {
	return &RelinearizationKey{EvaluationKey: *rlk.EvaluationKey.CopyNew()}
}

// GaloisKey is a type of [EvaluationKey] used to evaluate automorphisms on ciphertext.
// An automorphism pi: X^{i} -> X^{i*GaloisElement} changes the key under which the
// ciphertext is encrypted from s to pi(s). Thus, the ciphertext must be re-encrypted
// from pi(s) to s to ensure correctness, which is done with the corresponding GaloisKey.
//
// Lattigo implements automorphisms differently than the usual way (which is to first
// apply the automorphism and then the evaluation key). Instead the order of operations
// is reversed, the GaloisKey for pi^{-1} is evaluated on the ciphertext, outputting a
// ciphertext encrypted under pi^{-1}(s), and then the automorphism pi is applied. This
// enables a more efficient evaluation, by only having to apply the automorphism on the
// final result (instead of having to apply it on the decomposed ciphertext).
type GaloisKey struct {
	GaloisElement uint64
	NthRoot       uint64
	EvaluationKey
}

// NewGaloisKey allocates a new [GaloisKey] with zero coefficients and GaloisElement set to zero.
func NewGaloisKey(params ParameterProvider, evkParams ...EvaluationKeyParameters) *GaloisKey {
	p := *params.GetRLWEParameters()
	levelQ, levelP, BaseTwoDecomposition, compressed := ResolveEvaluationKeyParameters(p, evkParams)
	return newGaloisKey(p, levelQ, levelP, BaseTwoDecomposition, compressed)
}

func newGaloisKey(params Parameters, levelQ, levelP, BaseTwoDecomposition int, compressed bool) *GaloisKey {
	degree := 1
	if compressed {
		degree = 0
	}
	return &GaloisKey{
		EvaluationKey: EvaluationKey{
			GadgetCiphertext: *NewGadgetCiphertext(params, degree, levelQ, levelP, BaseTwoDecomposition),
		},
		NthRoot: params.GetRLWEParameters().RingQ().NthRoot(),
	}
}

// CopyNew creates a deep copy of the object and returns it
func (gk GaloisKey) CopyNew() *GaloisKey {
	return &GaloisKey{
		GaloisElement: gk.GaloisElement,
		NthRoot:       gk.NthRoot,
		EvaluationKey: *gk.EvaluationKey.CopyNew(),
	}
}

// BinarySize returns the serialized size of the object in bytes.
func (gk GaloisKey) BinarySize() (size int) {
	return gk.EvaluationKey.BinarySize() + 16
}

// WriteTo writes the object on an [io.Writer]. It implements the [io.WriterTo]
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the [buffer.Writer] interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a [bufio.Writer]. Since this requires allocations, it
// is preferable to pass a [buffer.Writer] directly:
//
//   - When writing multiple times to a [io.Writer], it is preferable to first wrap the
//     io.Writer in a pre-allocated [bufio.Writer].
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (gk GaloisKey) WriteTo(w io.Writer) (n int64, err error) {
	switch w := w.(type) {
	case buffer.Writer:

		var inc int64

		if inc, err = buffer.WriteUint64(w, gk.GaloisElement); err != nil {
			return n + inc, err
		}

		n += inc

		if inc, err = buffer.WriteUint64(w, gk.NthRoot); err != nil {
			return n + inc, err
		}

		n += inc

		if inc, err = gk.EvaluationKey.WriteTo(w); err != nil {
			return n + inc, err
		}

		n += inc

		return

	default:
		return gk.WriteTo(bufio.NewWriter(w))
	}
}

// ReadFrom reads on the object from an [io.Writer]. It implements the
// [io.ReaderFrom] interface.
//
// Unless r implements the [buffer.Reader] interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a [bufio.Reader]. Since this requires allocation, it
// is preferable to pass a [buffer.Reader] directly:
//
//   - When reading multiple values from a [io.Reader], it is preferable to first
//     first wrap io.Reader in a pre-allocated [bufio.Reader].
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (gk *GaloisKey) ReadFrom(r io.Reader) (n int64, err error) {
	switch r := r.(type) {
	case buffer.Reader:

		var inc int64

		if inc, err = buffer.ReadUint64(r, &gk.GaloisElement); err != nil {
			return n + inc, err
		}

		n += inc

		if inc, err = buffer.ReadUint64(r, &gk.NthRoot); err != nil {
			return n + inc, err
		}

		n += inc

		if inc, err = gk.EvaluationKey.ReadFrom(r); err != nil {
			return n + inc, err
		}

		n += inc

		return
	default:
		return gk.ReadFrom(bufio.NewReader(r))
	}
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (gk GaloisKey) MarshalBinary() (p []byte, err error) {
	buf := buffer.NewBufferSize(gk.BinarySize())
	_, err = gk.WriteTo(buf)
	return buf.Bytes(), err
}

// UnmarshalBinary decodes a slice of bytes generated by
// [GaloisKey.MarshalBinary] or [GaloisKey.WriteTo] on the object.
func (gk *GaloisKey) UnmarshalBinary(p []byte) (err error) {
	_, err = gk.ReadFrom(buffer.NewBuffer(p))
	return
}

// EvaluationKeySet is an interface implementing methods
// to load the [RelinearizationKey] and the [GaloisKey] in the [Evaluator].
// Implementations of this interface must be safe for concurrent use.
type EvaluationKeySet interface {

	// GetGaloisKey retrieves the Galois key for the automorphism X^{i} -> X^{i*galEl}.
	GetGaloisKey(galEl uint64) (evk *GaloisKey, err error)

	// GetGaloisKeysList returns the list of all the Galois elements
	// for which a Galois key exists in the object.
	GetGaloisKeysList() (galEls []uint64)

	// GetRelinearizationKey retrieves the RelinearizationKey.
	GetRelinearizationKey() (evk *RelinearizationKey, err error)

	// ShallowCopy returns a thread-safe copy of the underlying object.
	ShallowCopy() EvaluationKeySet
}

// MemEvaluationKeySet is a basic in-memory implementation of the [EvaluationKeySet] interface.
type MemEvaluationKeySet struct {
	RelinearizationKey *RelinearizationKey
	GaloisKeys         structs.Map[uint64, GaloisKey]
}

// NewMemEvaluationKeySet returns a new [EvaluationKeySet] with the provided [RelinearizationKey] and the [GaloisKey].
func NewMemEvaluationKeySet(relinKey *RelinearizationKey, galoisKeys ...*GaloisKey) (eks *MemEvaluationKeySet) {
	eks = &MemEvaluationKeySet{GaloisKeys: map[uint64]*GaloisKey{}}
	eks.RelinearizationKey = relinKey
	for _, k := range galoisKeys {
		eks.GaloisKeys[k.GaloisElement] = k
	}
	return eks
}

// GetGaloisKey retrieves the [GaloisKey] for the automorphism X^{i} -> X^{i*galEl}.
func (evk MemEvaluationKeySet) GetGaloisKey(galEl uint64) (gk *GaloisKey, err error) {
	var ok bool
	if gk, ok = evk.GaloisKeys[galEl]; !ok {
		return nil, fmt.Errorf("GaloisKey[%d] is nil", galEl)
	}

	return
}

// GetGaloisKeysList returns the list of all the Galois elements
// for which a Galois key exists in the object.
func (evk MemEvaluationKeySet) GetGaloisKeysList() (galEls []uint64) {

	if evk.GaloisKeys == nil {
		return []uint64{}
	}

	galEls = make([]uint64, len(evk.GaloisKeys))

	var i int
	for galEl := range evk.GaloisKeys {
		galEls[i] = galEl
		i++
	}

	return
}

// GetRelinearizationKey retrieves the [RelinearizationKey].
func (evk MemEvaluationKeySet) GetRelinearizationKey() (rk *RelinearizationKey, err error) {
	if evk.RelinearizationKey != nil {
		return evk.RelinearizationKey, nil
	}

	return nil, fmt.Errorf("RelinearizationKey is nil")
}

func (evk MemEvaluationKeySet) BinarySize() (size int) {

	size++
	if evk.RelinearizationKey != nil {
		size += evk.RelinearizationKey.BinarySize()
	}

	size++
	if evk.GaloisKeys != nil {
		size += evk.GaloisKeys.BinarySize()
	}

	return
}

// ShallowCopy returns a thread-safe copy of the MemEvaluationKey object.
func (evk *MemEvaluationKeySet) ShallowCopy() EvaluationKeySet {
	return evk
}

// WriteTo writes the object on an [io.Writer]. It implements the [io.WriterTo]
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the [buffer.Writer] interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a [bufio.Writer]. Since this requires allocations, it
// is preferable to pass a [buffer.Writer] directly:
//
//   - When writing multiple times to a [io.Writer], it is preferable to first wrap the
//     io.Writer in a pre-allocated [bufio.Writer].
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (evk MemEvaluationKeySet) WriteTo(w io.Writer) (n int64, err error) {
	switch w := w.(type) {
	case buffer.Writer:

		var inc int64

		if evk.RelinearizationKey != nil {
			if inc, err = buffer.WriteUint8(w, 1); err != nil {
				return inc, err
			}

			n += inc

			if inc, err = evk.RelinearizationKey.WriteTo(w); err != nil {
				return n + inc, err
			}

			n += inc

		} else {
			if inc, err = buffer.WriteUint8(w, 0); err != nil {
				return inc, err
			}
			n += inc
		}

		if evk.GaloisKeys != nil {
			if inc, err = buffer.WriteUint8(w, 1); err != nil {
				return inc, err
			}

			n += inc

			if inc, err = evk.GaloisKeys.WriteTo(w); err != nil {
				return n + inc, err
			}

			n += inc

		} else {
			if inc, err = buffer.WriteUint8(w, 0); err != nil {
				return inc, err
			}
			n += inc
		}

		return n, w.Flush()

	default:
		return evk.WriteTo(bufio.NewWriter(w))
	}
}

// ReadFrom reads on the object from an [io.Writer]. It implements the
// [io.ReaderFrom] interface.
//
// Unless r implements the [buffer.Reader] interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a [bufio.Reader]. Since this requires allocation, it
// is preferable to pass a [buffer.Reader] directly:
//
//   - When reading multiple values from a [io.Reader], it is preferable to first
//     first wrap io.Reader in a pre-allocated [bufio.Reader].
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (evk *MemEvaluationKeySet) ReadFrom(r io.Reader) (n int64, err error) {
	switch r := r.(type) {
	case buffer.Reader:

		var inc int64

		var hasKey uint8

		if inc, err = buffer.ReadUint8(r, &hasKey); err != nil {
			return inc, err
		}

		n += inc

		if hasKey == 1 {

			if evk.RelinearizationKey == nil {
				evk.RelinearizationKey = new(RelinearizationKey)
			}

			if inc, err = evk.RelinearizationKey.ReadFrom(r); err != nil {
				return n + inc, err
			}

			n += inc
		}

		if inc, err = buffer.ReadUint8(r, &hasKey); err != nil {
			return inc, err
		}

		n += inc

		if hasKey == 1 {

			if evk.GaloisKeys == nil {
				evk.GaloisKeys = structs.Map[uint64, GaloisKey]{}
			}

			if inc, err = evk.GaloisKeys.ReadFrom(r); err != nil {
				return n + inc, err
			}

			n += inc
		}

		return n, nil

	default:
		return evk.ReadFrom(bufio.NewReader(r))
	}
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (evk MemEvaluationKeySet) MarshalBinary() (p []byte, err error) {
	buf := buffer.NewBufferSize(evk.BinarySize())
	_, err = evk.WriteTo(buf)
	return buf.Bytes(), err
}

// UnmarshalBinary decodes a slice of bytes generated by
// [MemEvaluationKeySet.MarshalBinary] or [MemEvaluationKeySet.WriteTo] on the object.
func (evk *MemEvaluationKeySet) UnmarshalBinary(p []byte) (err error) {
	_, err = evk.ReadFrom(buffer.NewBuffer(p))
	return
}
