package ckks

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/utils/bignum"
)

// SimpleBootstrapper is an implementation of the rlwe.Bootstrapping interface that
// uses the secret-key to decrypt and re-encrypt the bootstrapped ciphertext.
type SimpleBootstrapper struct {
	Parameters
	*Encoder
	rlwe.Decryptor
	rlwe.Encryptor
	sk      *rlwe.SecretKey
	Values  []*bignum.Complex
	Counter int // records the number of bootstrapping
}

func NewSimpleBootstrapper(params Parameters, sk *rlwe.SecretKey) rlwe.Bootstrapper {
	return &SimpleBootstrapper{
		params,
		NewEncoder(params),
		NewDecryptor(params, sk),
		NewEncryptor(params, sk),
		sk,
		make([]*bignum.Complex, params.N()),
		0}
}

func (d *SimpleBootstrapper) Bootstrap(ct *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	values := d.Values[:1<<ct.LogSlots]
	if err := d.Decode(d.DecryptNew(ct), values); err != nil {
		return nil, err
	}
	pt := NewPlaintext(d.Parameters, d.MaxLevel())
	pt.MetaData = ct.MetaData
	if err := d.Encode(values, pt); err != nil {
		return nil, err
	}
	ct.Resize(1, d.MaxLevel())
	d.Encrypt(pt, ct)
	d.Counter++
	return ct, nil
}

func (d *SimpleBootstrapper) BootstrapMany(cts []*rlwe.Ciphertext) ([]*rlwe.Ciphertext, error) {
	for i := range cts {
		cts[i], _ = d.Bootstrap(cts[i])
	}
	return cts, nil
}

func (d *SimpleBootstrapper) Depth() int {
	return 0
}

func (d *SimpleBootstrapper) MinimumInputLevel() int {
	return 0
}

func (d *SimpleBootstrapper) OuputLevel() int {
	return d.MaxLevel()
}
