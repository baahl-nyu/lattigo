package ring

import (
	"bufio"
	"io"

	"github.com/baahl-nyu/lattigo/v6/utils"
	"github.com/baahl-nyu/lattigo/v6/utils/buffer"
	"github.com/baahl-nyu/lattigo/v6/utils/structs"
)

// Poly is the structure that contains the coefficients of a polynomial.
type Poly struct {
	Coeffs structs.Matrix[uint64]
}

// NewPoly creates a new polynomial with N coefficients set to zero and Level+1 moduli.
func NewPoly(N, Level int) (pol Poly) {
	Coeffs := make([][]uint64, Level+1)
	for i := range Coeffs {
		Coeffs[i] = make([]uint64, N)
	}
	return Poly{Coeffs: Coeffs}
}

// Resize resizes the level of the target polynomial to the provided level.
// If the provided level is larger than the current level, then allocates zero
// coefficients, otherwise dereferences the coefficients above the provided level.
func (pol *Poly) Resize(level int) {
	N := pol.N()
	if pol.Level() > level {
		pol.Coeffs = pol.Coeffs[:level+1]
	} else if level > pol.Level() {
		prevLevel := pol.Level()
		pol.Coeffs = append(pol.Coeffs, make([][]uint64, level-prevLevel)...)
		for i := prevLevel + 1; i < level+1; i++ {
			pol.Coeffs[i] = make([]uint64, N)
		}
	}
}

// N returns the number of coefficients of the polynomial, which equals the degree of the Ring cyclotomic polynomial.
func (pol Poly) N() int {
	if len(pol.Coeffs) == 0 {
		return 0
	}
	return len(pol.Coeffs[0])
}

// Level returns the current number of moduli minus 1.
func (pol Poly) Level() int {
	return len(pol.Coeffs) - 1
}

// Zero sets all coefficients of the target polynomial to 0.
func (pol Poly) Zero() {
	for i := range pol.Coeffs {
		ZeroVec(pol.Coeffs[i])
	}
}

// CopyNew creates an exact copy of the target polynomial.
func (pol Poly) CopyNew() *Poly {
	return &Poly{
		Coeffs: pol.Coeffs.CopyNew(),
	}
}

// Copy copies the coefficients of p1 on the target polynomial.
// This method does nothing if the underlying arrays are the same.
// This method will resize the target polynomial to the level of
// the input polynomial.
func (pol *Poly) Copy(p1 Poly) {
	pol.Resize(p1.Level())
	pol.CopyLvl(p1.Level(), p1)
}

// CopyLvl copies the coefficients of p1 on the target polynomial.
// This method does nothing if the underlying arrays are the same.
// Expects the degree of both polynomials to be identical.
func (pol *Poly) CopyLvl(level int, p1 Poly) {
	for i := 0; i < level+1; i++ {
		if !utils.Alias1D(pol.Coeffs[i], p1.Coeffs[i]) {
			copy(pol.Coeffs[i], p1.Coeffs[i])
		}
	}
}

// Equal returns true if the receiver Poly is equal to the provided other Poly.
// This function checks for strict equality between the polynomial coefficients
// (i.e., it does not consider congruence as equality within the ring like
// `Ring.Equal` does).
func (pol Poly) Equal(other *Poly) bool {
	return pol.Coeffs.Equal(other.Coeffs)
}

// BinarySize returns the serialized size of the object in bytes.
func (pol Poly) BinarySize() (size int) {
	return pol.Coeffs.BinarySize()
}

// WriteTo writes the object on an io.Writer. It implements the io.WriterTo
// interface, and will write exactly object.BinarySize() bytes on w.
//
// Unless w implements the buffer.Writer interface (see lattigo/utils/buffer/writer.go),
// it will be wrapped into a bufio.Writer. Since this requires allocations, it
// is preferable to pass a buffer.Writer directly:
//
//   - When writing multiple times to a io.Writer, it is preferable to first wrap the
//     io.Writer in a pre-allocated bufio.Writer.
//   - When writing to a pre-allocated var b []byte, it is preferable to pass
//     buffer.NewBuffer(b) as w (see lattigo/utils/buffer/buffer.go).
func (pol Poly) WriteTo(w io.Writer) (n int64, err error) {
	switch w := w.(type) {
	case buffer.Writer:
		if n, err = pol.Coeffs.WriteTo(w); err != nil {
			return
		}
		return n, w.Flush()
	default:
		return pol.WriteTo(bufio.NewWriter(w))
	}
}

// ReadFrom reads on the object from an io.Writer. It implements the
// io.ReaderFrom interface.
//
// Unless r implements the buffer.Reader interface (see see lattigo/utils/buffer/reader.go),
// it will be wrapped into a bufio.Reader. Since this requires allocation, it
// is preferable to pass a buffer.Reader directly:
//
//   - When reading multiple values from a io.Reader, it is preferable to first
//     first wrap io.Reader in a pre-allocated bufio.Reader.
//   - When reading from a var b []byte, it is preferable to pass a buffer.NewBuffer(b)
//     as w (see lattigo/utils/buffer/buffer.go).
func (pol *Poly) ReadFrom(r io.Reader) (n int64, err error) {
	switch r := r.(type) {
	case buffer.Reader:
		if n, err = pol.Coeffs.ReadFrom(r); err != nil {
			return
		}
		return n, nil
	default:
		return pol.ReadFrom(bufio.NewReader(r))
	}
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (pol Poly) MarshalBinary() (p []byte, err error) {
	buf := buffer.NewBufferSize(pol.BinarySize())
	_, err = pol.WriteTo(buf)
	return buf.Bytes(), err
}

// UnmarshalBinary decodes a slice of bytes generated by
// MarshalBinary or WriteTo on the object.
func (pol *Poly) UnmarshalBinary(p []byte) (err error) {
	_, err = pol.ReadFrom(buffer.NewBuffer(p))
	return
}
