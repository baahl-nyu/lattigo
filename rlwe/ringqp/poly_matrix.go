package ringqp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tuneinsight/lattigo/v4/utils/buffer"
)

// PolyMatrix is a struct storing a vector of PolyVector.
type PolyMatrix []*PolyVector

// NewPolyMatrix allocates a new PolyMatrix of size rows x cols.
func NewPolyMatrix(N, levelQ, levelP, rows, cols int) *PolyMatrix {
	m := make([]*PolyVector, rows)

	for i := range m {
		m[i] = NewPolyVector(N, levelQ, levelP, cols)
	}

	pm := PolyMatrix(m)

	return &pm
}

// Set sets a poly matrix to the double slice of *Poly.
// Overwrites the current states of the poly matrix.
func (pm *PolyMatrix) Set(polys [][]Poly) {

	m := PolyMatrix(make([]*PolyVector, len(polys)))
	for i := range m {
		m[i] = new(PolyVector)
		m[i].Set(polys[i])
	}

	*pm = m
}

// Get returns the underlying double slice of *Poly.
func (pm *PolyMatrix) Get() [][]Poly {
	m := *pm
	polys := make([][]Poly, len(m))
	for i := range polys {
		polys[i] = m[i].Get()
	}
	return polys
}

// N returns the ring degree of the first polynomial in the matrix of polynomials.
func (pm *PolyMatrix) N() int {
	return (*pm)[0].N()
}

// LevelQ returns the LevelQ of the first polynomial in the matrix of polynomials.
func (pm *PolyMatrix) LevelQ() int {
	return (*pm)[0].LevelP()
}

// LevelP returns the LevelP of the first polynomial in the matrix of polynomials.
func (pm *PolyMatrix) LevelP() int {
	return (*pm)[0].LevelP()
}

// Resize resizes the level, rows and columns of the matrix of polynomials, allocating if necessary.
func (pm *PolyMatrix) Resize(levelQ, levelP, rows, cols int) {
	N := pm.N()

	v := *pm

	for i := range v {
		v[i].Resize(levelQ, levelP, cols)
	}

	if len(v) > rows {
		v = v[:rows+1]
	} else {
		for i := len(v); i < rows+1; i++ {
			v = append(v, NewPolyVector(N, levelQ, levelP, cols))
		}
	}

	*pm = v
}

// BinarySize returns the size in bytes of the object
// when encoded using MarshalBinary, Read or WriteTo.
func (pm *PolyMatrix) BinarySize() (size int) {
	size += 8
	for _, m := range *pm {
		size += m.BinarySize()
	}
	return
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (pm *PolyMatrix) MarshalBinary() (p []byte, err error) {
	p = make([]byte, pm.BinarySize())
	_, err = pm.Read(p)
	return
}

// Read encodes the object into a binary form on a preallocated slice of bytes
// and returns the number of bytes written.
func (pm *PolyMatrix) Read(b []byte) (n int, err error) {

	m := *pm

	binary.LittleEndian.PutUint64(b[n:], uint64(len(m)))
	n += 8

	var inc int
	for i := range m {
		if inc, err = m[i].Read(b[n:]); err != nil {
			return n + inc, err
		}

		n += inc
	}

	return
}

// WriteTo writes the object on an io.Writer.
// To ensure optimal efficiency and minimal allocations, the user is encouraged
// to provide a struct implementing the interface buffer.Writer, which defines
// a subset of the method of the bufio.Writer.
// If w is not compliant to the buffer.Writer interface, it will be wrapped in
// a new bufio.Writer.
// For additional information, see lattigo/utils/buffer/writer.go.
func (pm *PolyMatrix) WriteTo(w io.Writer) (int64, error) {
	switch w := w.(type) {
	case buffer.Writer:

		var err error
		var n int64

		m := *pm

		var inc int
		if inc, err = buffer.WriteInt(w, len(m)); err != nil {
			return int64(inc), err
		}

		n += int64(inc)

		for i := range m {
			var inc int64
			if inc, err = m[i].WriteTo(w); err != nil {
				return n + inc, err
			}

			n += inc
		}

		return n, nil

	default:
		return pm.WriteTo(bufio.NewWriter(w))
	}
}

// UnmarshalBinary decodes a slice of bytes generated by MarshalBinary
// or Read on the object.
func (pm *PolyMatrix) UnmarshalBinary(p []byte) (err error) {
	_, err = pm.Write(p)
	return
}

// Write decodes a slice of bytes generated by MarshalBinary, WriteTo or
// Read on the object and returns the number of bytes read.
func (pm *PolyMatrix) Write(p []byte) (n int, err error) {
	size := int(binary.LittleEndian.Uint64(p[n:]))
	n += 8

	if len(*pm) != size {
		*pm = make([]*PolyVector, size)
	}

	m := *pm

	var inc int
	for i := range m {
		if m[i] == nil {
			m[i] = new(PolyVector)
		}

		if inc, err = m[i].Write(p[n:]); err != nil {
			return n + inc, err
		}

		n += inc
	}

	return
}

// ReadFrom reads on the object from an io.Writer.
// To ensure optimal efficiency and minimal allocations, the user is encouraged
// to provide a struct implementing the interface buffer.Reader, which defines
// a subset of the method of the bufio.Reader.
// If r is not compliant to the buffer.Reader interface, it will be wrapped in
// a new bufio.Reader.
// For additional information, see lattigo/utils/buffer/reader.go.
func (pm *PolyMatrix) ReadFrom(r io.Reader) (int64, error) {
	switch r := r.(type) {
	case buffer.Reader:

		var err error
		var size, n int

		if n, err = buffer.ReadInt(r, &size); err != nil {
			return int64(n), fmt.Errorf("cannot ReadFrom: size: %w", err)
		}

		if len(*pm) != size {
			*pm = make([]*PolyVector, size)
		}

		m := *pm

		for i := range m {

			if m[i] == nil {
				m[i] = new(PolyVector)
			}

			var inc int64
			if inc, err = m[i].ReadFrom(r); err != nil {
				return int64(n) + inc, err
			}

			n += int(inc)
		}

		return int64(n), nil

	default:
		return pm.ReadFrom(bufio.NewReader(r))
	}
}
