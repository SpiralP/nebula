package cert

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/slackhq/nebula/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/protobuf/proto"
)

func TestMarshalingNebulaCertificate(t *testing.T) {
	before := time.Now().Add(time.Second * -60).Round(time.Second)
	after := time.Now().Add(time.Second * 60).Round(time.Second)
	pubKey := []byte("1234567890abcedfghij1234567890ab")

	nc := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "testing",
			Ips: []*net.IPNet{
				{IP: net.ParseIP("10.1.1.1"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("10.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
				{IP: net.ParseIP("10.1.1.3"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
			},
			Subnets: []*net.IPNet{
				{IP: net.ParseIP("9.1.1.1"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
				{IP: net.ParseIP("9.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("9.1.1.3"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
			},
			Groups:    []string{"test-group1", "test-group2", "test-group3"},
			NotBefore: before,
			NotAfter:  after,
			PublicKey: pubKey,
			IsCA:      false,
			Issuer:    "1234567890abcedfghij1234567890ab",
		},
		Signature: []byte("1234567890abcedfghij1234567890ab"),
	}

	b, err := nc.Marshal()
	assert.Nil(t, err)
	//t.Log("Cert size:", len(b))

	nc2, err := UnmarshalNebulaCertificate(b)
	assert.Nil(t, err)

	assert.Equal(t, nc.Signature, nc2.Signature)
	assert.Equal(t, nc.Details.Name, nc2.Details.Name)
	assert.Equal(t, nc.Details.NotBefore, nc2.Details.NotBefore)
	assert.Equal(t, nc.Details.NotAfter, nc2.Details.NotAfter)
	assert.Equal(t, nc.Details.PublicKey, nc2.Details.PublicKey)
	assert.Equal(t, nc.Details.IsCA, nc2.Details.IsCA)

	// IP byte arrays can be 4 or 16 in length so we have to go this route
	assert.Equal(t, len(nc.Details.Ips), len(nc2.Details.Ips))
	for i, wIp := range nc.Details.Ips {
		assert.Equal(t, wIp.String(), nc2.Details.Ips[i].String())
	}

	assert.Equal(t, len(nc.Details.Subnets), len(nc2.Details.Subnets))
	for i, wIp := range nc.Details.Subnets {
		assert.Equal(t, wIp.String(), nc2.Details.Subnets[i].String())
	}

	assert.EqualValues(t, nc.Details.Groups, nc2.Details.Groups)
}

func TestNebulaCertificate_Sign(t *testing.T) {
	before := time.Now().Add(time.Second * -60).Round(time.Second)
	after := time.Now().Add(time.Second * 60).Round(time.Second)
	pubKey := []byte("1234567890abcedfghij1234567890ab")

	nc := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "testing",
			Ips: []*net.IPNet{
				{IP: net.ParseIP("10.1.1.1"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("10.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
				{IP: net.ParseIP("10.1.1.3"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
			},
			Subnets: []*net.IPNet{
				{IP: net.ParseIP("9.1.1.1"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
				{IP: net.ParseIP("9.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("9.1.1.3"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
			},
			Groups:    []string{"test-group1", "test-group2", "test-group3"},
			NotBefore: before,
			NotAfter:  after,
			PublicKey: pubKey,
			IsCA:      false,
			Issuer:    "1234567890abcedfghij1234567890ab",
		},
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(t, err)
	assert.False(t, nc.CheckSignature(pub))
	assert.Nil(t, nc.Sign(Curve_CURVE25519, priv))
	assert.True(t, nc.CheckSignature(pub))

	_, err = nc.Marshal()
	assert.Nil(t, err)
	//t.Log("Cert size:", len(b))
}

func TestNebulaCertificate_SignP256(t *testing.T) {
	before := time.Now().Add(time.Second * -60).Round(time.Second)
	after := time.Now().Add(time.Second * 60).Round(time.Second)
	pubKey := []byte("01234567890abcedfghij1234567890ab1234567890abcedfghij1234567890ab")

	nc := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "testing",
			Ips: []*net.IPNet{
				{IP: net.ParseIP("10.1.1.1"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("10.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
				{IP: net.ParseIP("10.1.1.3"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
			},
			Subnets: []*net.IPNet{
				{IP: net.ParseIP("9.1.1.1"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
				{IP: net.ParseIP("9.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("9.1.1.3"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
			},
			Groups:    []string{"test-group1", "test-group2", "test-group3"},
			NotBefore: before,
			NotAfter:  after,
			PublicKey: pubKey,
			IsCA:      false,
			Curve:     Curve_P256,
			Issuer:    "1234567890abcedfghij1234567890ab",
		},
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pub := elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y)
	rawPriv := priv.D.FillBytes(make([]byte, 32))

	assert.Nil(t, err)
	assert.False(t, nc.CheckSignature(pub))
	assert.Nil(t, nc.Sign(Curve_P256, rawPriv))
	assert.True(t, nc.CheckSignature(pub))

	_, err = nc.Marshal()
	assert.Nil(t, err)
	//t.Log("Cert size:", len(b))
}

func TestNebulaCertificate_Expired(t *testing.T) {
	nc := NebulaCertificate{
		Details: NebulaCertificateDetails{
			NotBefore: time.Now().Add(time.Second * -60).Round(time.Second),
			NotAfter:  time.Now().Add(time.Second * 60).Round(time.Second),
		},
	}

	assert.True(t, nc.Expired(time.Now().Add(time.Hour)))
	assert.True(t, nc.Expired(time.Now().Add(-time.Hour)))
	assert.False(t, nc.Expired(time.Now()))
}

func TestNebulaCertificate_MarshalJSON(t *testing.T) {
	time.Local = time.UTC
	pubKey := []byte("1234567890abcedfghij1234567890ab")

	nc := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "testing",
			Ips: []*net.IPNet{
				{IP: net.ParseIP("10.1.1.1"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("10.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
				{IP: net.ParseIP("10.1.1.3"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
			},
			Subnets: []*net.IPNet{
				{IP: net.ParseIP("9.1.1.1"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
				{IP: net.ParseIP("9.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("9.1.1.3"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
			},
			Groups:    []string{"test-group1", "test-group2", "test-group3"},
			NotBefore: time.Date(1, 0, 0, 1, 0, 0, 0, time.UTC),
			NotAfter:  time.Date(1, 0, 0, 2, 0, 0, 0, time.UTC),
			PublicKey: pubKey,
			IsCA:      false,
			Issuer:    "1234567890abcedfghij1234567890ab",
		},
		Signature: []byte("1234567890abcedfghij1234567890ab"),
	}

	b, err := nc.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(
		t,
		"{\"details\":{\"curve\":\"CURVE25519\",\"groups\":[\"test-group1\",\"test-group2\",\"test-group3\"],\"ips\":[\"10.1.1.1/24\",\"10.1.1.2/16\",\"10.1.1.3/ff00ff00\"],\"isCa\":false,\"issuer\":\"1234567890abcedfghij1234567890ab\",\"name\":\"testing\",\"notAfter\":\"0000-11-30T02:00:00Z\",\"notBefore\":\"0000-11-30T01:00:00Z\",\"publicKey\":\"313233343536373839306162636564666768696a313233343536373839306162\",\"subnets\":[\"9.1.1.1/ff00ff00\",\"9.1.1.2/24\",\"9.1.1.3/16\"]},\"fingerprint\":\"26cb1c30ad7872c804c166b5150fa372f437aa3856b04edb4334b4470ec728e4\",\"signature\":\"313233343536373839306162636564666768696a313233343536373839306162\"}",
		string(b),
	)
}

func TestNebulaCertificate_Verify(t *testing.T) {
	ca, _, caKey, err := newTestCaCert(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)

	c, _, _, err := newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)

	h, err := ca.Sha256Sum()
	assert.Nil(t, err)

	caPool := NewCAPool()
	caPool.CAs[h] = ca

	f, err := c.Sha256Sum()
	assert.Nil(t, err)
	caPool.BlocklistFingerprint(f)

	v, err := c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate is in the block list")

	caPool.ResetCertBlocklist()
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	v, err = c.Verify(time.Now().Add(time.Hour*1000), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "root certificate is expired")

	c, _, _, err = newTestCert(ca, caKey, time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now().Add(time.Minute*6), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate is expired")

	// Test group assertion
	ca, _, caKey, err = newTestCaCert(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{"test1", "test2"})
	assert.Nil(t, err)

	caPem, err := ca.MarshalToPEM()
	assert.Nil(t, err)

	caPool = NewCAPool()
	caPool.AddCACertificate(caPem)

	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{"test1", "bad"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained a group not present on the signing ca: bad")

	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{"test1"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)
}

func TestNebulaCertificate_VerifyP256(t *testing.T) {
	ca, _, caKey, err := newTestCaCertP256(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)

	c, _, _, err := newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)

	h, err := ca.Sha256Sum()
	assert.Nil(t, err)

	caPool := NewCAPool()
	caPool.CAs[h] = ca

	f, err := c.Sha256Sum()
	assert.Nil(t, err)
	caPool.BlocklistFingerprint(f)

	v, err := c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate is in the block list")

	caPool.ResetCertBlocklist()
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	v, err = c.Verify(time.Now().Add(time.Hour*1000), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "root certificate is expired")

	c, _, _, err = newTestCert(ca, caKey, time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now().Add(time.Minute*6), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate is expired")

	// Test group assertion
	ca, _, caKey, err = newTestCaCertP256(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{"test1", "test2"})
	assert.Nil(t, err)

	caPem, err := ca.MarshalToPEM()
	assert.Nil(t, err)

	caPool = NewCAPool()
	caPool.AddCACertificate(caPem)

	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{"test1", "bad"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained a group not present on the signing ca: bad")

	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{"test1"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)
}

func TestNebulaCertificate_Verify_IPs(t *testing.T) {
	_, caIp1, _ := net.ParseCIDR("10.0.0.0/16")
	_, caIp2, _ := net.ParseCIDR("192.168.0.0/24")
	ca, _, caKey, err := newTestCaCert(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{caIp1, caIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)

	caPem, err := ca.MarshalToPEM()
	assert.Nil(t, err)

	caPool := NewCAPool()
	caPool.AddCACertificate(caPem)

	// ip is outside the network
	cIp1 := &net.IPNet{IP: net.ParseIP("10.1.0.0"), Mask: []byte{255, 255, 255, 0}}
	cIp2 := &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 0, 0}}
	c, _, _, err := newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{cIp1, cIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err := c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained an ip assignment outside the limitations of the signing ca: 10.1.0.0/24")

	// ip is outside the network reversed order of above
	cIp1 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("10.1.0.0"), Mask: []byte{255, 255, 255, 0}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{cIp1, cIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained an ip assignment outside the limitations of the signing ca: 10.1.0.0/24")

	// ip is within the network but mask is outside
	cIp1 = &net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: []byte{255, 254, 0, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 0}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{cIp1, cIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained an ip assignment outside the limitations of the signing ca: 10.0.1.0/15")

	// ip is within the network but mask is outside reversed order of above
	cIp1 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: []byte{255, 254, 0, 0}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{cIp1, cIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained an ip assignment outside the limitations of the signing ca: 10.0.1.0/15")

	// ip and mask are within the network
	cIp1 = &net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: []byte{255, 255, 0, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 128}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{cIp1, cIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	// Exact matches
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{caIp1, caIp2}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	// Exact matches reversed
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{caIp2, caIp1}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	// Exact matches reversed with just 1
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{caIp1}, []*net.IPNet{}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)
}

func TestNebulaCertificate_Verify_Subnets(t *testing.T) {
	_, caIp1, _ := net.ParseCIDR("10.0.0.0/16")
	_, caIp2, _ := net.ParseCIDR("192.168.0.0/24")
	ca, _, caKey, err := newTestCaCert(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{}, []*net.IPNet{caIp1, caIp2}, []string{"test"})
	assert.Nil(t, err)

	caPem, err := ca.MarshalToPEM()
	assert.Nil(t, err)

	caPool := NewCAPool()
	caPool.AddCACertificate(caPem)

	// ip is outside the network
	cIp1 := &net.IPNet{IP: net.ParseIP("10.1.0.0"), Mask: []byte{255, 255, 255, 0}}
	cIp2 := &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 0, 0}}
	c, _, _, err := newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{cIp1, cIp2}, []string{"test"})
	assert.Nil(t, err)
	v, err := c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained a subnet assignment outside the limitations of the signing ca: 10.1.0.0/24")

	// ip is outside the network reversed order of above
	cIp1 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("10.1.0.0"), Mask: []byte{255, 255, 255, 0}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{cIp1, cIp2}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained a subnet assignment outside the limitations of the signing ca: 10.1.0.0/24")

	// ip is within the network but mask is outside
	cIp1 = &net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: []byte{255, 254, 0, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 0}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{cIp1, cIp2}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained a subnet assignment outside the limitations of the signing ca: 10.0.1.0/15")

	// ip is within the network but mask is outside reversed order of above
	cIp1 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: []byte{255, 254, 0, 0}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{cIp1, cIp2}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.False(t, v)
	assert.EqualError(t, err, "certificate contained a subnet assignment outside the limitations of the signing ca: 10.0.1.0/15")

	// ip and mask are within the network
	cIp1 = &net.IPNet{IP: net.ParseIP("10.0.1.0"), Mask: []byte{255, 255, 0, 0}}
	cIp2 = &net.IPNet{IP: net.ParseIP("192.168.0.1"), Mask: []byte{255, 255, 255, 128}}
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{cIp1, cIp2}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	// Exact matches
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{caIp1, caIp2}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	// Exact matches reversed
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{caIp2, caIp1}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)

	// Exact matches reversed with just 1
	c, _, _, err = newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{caIp1}, []string{"test"})
	assert.Nil(t, err)
	v, err = c.Verify(time.Now(), caPool)
	assert.True(t, v)
	assert.Nil(t, err)
}

func TestNebulaCertificate_VerifyPrivateKey(t *testing.T) {
	ca, _, caKey, err := newTestCaCert(time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	err = ca.VerifyPrivateKey(Curve_CURVE25519, caKey)
	assert.Nil(t, err)

	_, _, caKey2, err := newTestCaCert(time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	err = ca.VerifyPrivateKey(Curve_CURVE25519, caKey2)
	assert.NotNil(t, err)

	c, _, priv, err := newTestCert(ca, caKey, time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	err = c.VerifyPrivateKey(Curve_CURVE25519, priv)
	assert.Nil(t, err)

	_, priv2 := x25519Keypair()
	err = c.VerifyPrivateKey(Curve_CURVE25519, priv2)
	assert.NotNil(t, err)
}

func TestNebulaCertificate_VerifyPrivateKeyP256(t *testing.T) {
	ca, _, caKey, err := newTestCaCertP256(time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	err = ca.VerifyPrivateKey(Curve_P256, caKey)
	assert.Nil(t, err)

	_, _, caKey2, err := newTestCaCertP256(time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	err = ca.VerifyPrivateKey(Curve_P256, caKey2)
	assert.NotNil(t, err)

	c, _, priv, err := newTestCert(ca, caKey, time.Time{}, time.Time{}, []*net.IPNet{}, []*net.IPNet{}, []string{})
	err = c.VerifyPrivateKey(Curve_P256, priv)
	assert.Nil(t, err)

	_, priv2 := p256Keypair()
	err = c.VerifyPrivateKey(Curve_P256, priv2)
	assert.NotNil(t, err)
}

func TestNewCAPoolFromBytes(t *testing.T) {
	noNewLines := `
# Current provisional, Remove once everything moves over to the real root.
-----BEGIN NEBULA CERTIFICATE-----
CkAKDm5lYnVsYSByb290IGNhKJfap9AFMJfg1+YGOiCUQGByMuNRhIlQBOyzXWbL
vcKBwDhov900phEfJ5DN3kABEkDCq5R8qBiu8sl54yVfgRcQXEDt3cHr8UTSLszv
bzBEr00kERQxxTzTsH8cpYEgRoipvmExvg8WP8NdAJEYJosB
-----END NEBULA CERTIFICATE-----
# root-ca01
-----BEGIN NEBULA CERTIFICATE-----
CkMKEW5lYnVsYSByb290IGNhIDAxKJL2u9EFMJL86+cGOiDPXMH4oU6HZTk/CqTG
BVG+oJpAoqokUBbI4U0N8CSfpUABEkB/Pm5A2xyH/nc8mg/wvGUWG3pZ7nHzaDMf
8/phAUt+FLzqTECzQKisYswKvE3pl9mbEYKbOdIHrxdIp95mo4sF
-----END NEBULA CERTIFICATE-----
`

	withNewLines := `
# Current provisional, Remove once everything moves over to the real root.

-----BEGIN NEBULA CERTIFICATE-----
CkAKDm5lYnVsYSByb290IGNhKJfap9AFMJfg1+YGOiCUQGByMuNRhIlQBOyzXWbL
vcKBwDhov900phEfJ5DN3kABEkDCq5R8qBiu8sl54yVfgRcQXEDt3cHr8UTSLszv
bzBEr00kERQxxTzTsH8cpYEgRoipvmExvg8WP8NdAJEYJosB
-----END NEBULA CERTIFICATE-----

# root-ca01


-----BEGIN NEBULA CERTIFICATE-----
CkMKEW5lYnVsYSByb290IGNhIDAxKJL2u9EFMJL86+cGOiDPXMH4oU6HZTk/CqTG
BVG+oJpAoqokUBbI4U0N8CSfpUABEkB/Pm5A2xyH/nc8mg/wvGUWG3pZ7nHzaDMf
8/phAUt+FLzqTECzQKisYswKvE3pl9mbEYKbOdIHrxdIp95mo4sF
-----END NEBULA CERTIFICATE-----

`

	expired := `
# expired certificate
-----BEGIN NEBULA CERTIFICATE-----
CjkKB2V4cGlyZWQouPmWjQYwufmWjQY6ILCRaoCkJlqHgv5jfDN4lzLHBvDzaQm4
vZxfu144hmgjQAESQG4qlnZi8DncvD/LDZnLgJHOaX1DWCHHEh59epVsC+BNgTie
WH1M9n4O7cFtGlM6sJJOS+rCVVEJ3ABS7+MPdQs=
-----END NEBULA CERTIFICATE-----
`

	p256 := `
# p256 certificate
-----BEGIN NEBULA CERTIFICATE-----
CmYKEG5lYnVsYSBQMjU2IHRlc3Qo4s+7mgYw4tXrsAc6QQRkaW2jFmllYvN4+/k2
6tctO9sPT3jOx8ES6M1nIqOhpTmZeabF/4rELDqPV4aH5jfJut798DUXql0FlF8H
76gvQAGgBgESRzBFAiEAib0/te6eMiZOKD8gdDeloMTS0wGuX2t0C7TFdUhAQzgC
IBNWYMep3ysx9zCgknfG5dKtwGTaqF++BWKDYdyl34KX
-----END NEBULA CERTIFICATE-----
`

	v2 := `
# valid PEM with the V2 header
-----BEGIN NEBULA CERTIFICATE V2-----
CmYKEG5lYnVsYSBQMjU2IHRlc3Qo4s+7mgYw4tXrsAc6QQRkaW2jFmllYvN4+/k2
-----END NEBULA CERTIFICATE V2-----
`

	rootCA := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "nebula root ca",
		},
	}

	rootCA01 := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "nebula root ca 01",
		},
	}

	rootCAP256 := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "nebula P256 test",
		},
	}

	p, warn, err := NewCAPoolFromBytes([]byte(noNewLines))
	assert.Nil(t, err)
	assert.Nil(t, warn)
	assert.Equal(t, p.CAs[string("c9bfaf7ce8e84b2eeda2e27b469f4b9617bde192efd214b68891ecda6ed49522")].Details.Name, rootCA.Details.Name)
	assert.Equal(t, p.CAs[string("5c9c3f23e7ee7fe97637cbd3a0a5b854154d1d9aaaf7b566a51f4a88f76b64cd")].Details.Name, rootCA01.Details.Name)

	pp, warn, err := NewCAPoolFromBytes([]byte(withNewLines))
	assert.Nil(t, err)
	assert.Nil(t, warn)
	assert.Equal(t, pp.CAs[string("c9bfaf7ce8e84b2eeda2e27b469f4b9617bde192efd214b68891ecda6ed49522")].Details.Name, rootCA.Details.Name)
	assert.Equal(t, pp.CAs[string("5c9c3f23e7ee7fe97637cbd3a0a5b854154d1d9aaaf7b566a51f4a88f76b64cd")].Details.Name, rootCA01.Details.Name)

	// expired cert, no valid certs
	ppp, warn, err := NewCAPoolFromBytes([]byte(expired))
	assert.Error(t, err, "no valid CA certificates present")
	assert.Len(t, warn, 1)
	assert.Error(t, warn[0], ErrExpired)
	assert.Nil(t, ppp)

	// expired cert, with valid certs
	pppp, warn, err := NewCAPoolFromBytes(append([]byte(expired), noNewLines...))
	assert.Len(t, warn, 1)
	assert.Nil(t, err)
	assert.Error(t, warn[0], ErrExpired)
	assert.Equal(t, pppp.CAs[string("c9bfaf7ce8e84b2eeda2e27b469f4b9617bde192efd214b68891ecda6ed49522")].Details.Name, rootCA.Details.Name)
	assert.Equal(t, pppp.CAs[string("5c9c3f23e7ee7fe97637cbd3a0a5b854154d1d9aaaf7b566a51f4a88f76b64cd")].Details.Name, rootCA01.Details.Name)
	assert.Equal(t, pppp.CAs[string("152070be6bb19bc9e3bde4c2f0e7d8f4ff5448b4c9856b8eccb314fade0229b0")].Details.Name, "expired")
	assert.Equal(t, len(pppp.CAs), 3)

	ppppp, warn, err := NewCAPoolFromBytes([]byte(p256))
	assert.Nil(t, err)
	assert.Nil(t, warn)
	assert.Equal(t, ppppp.CAs[string("a7938893ec8c4ef769b06d7f425e5e46f7a7f5ffa49c3bcf4a86b608caba9159")].Details.Name, rootCAP256.Details.Name)
	assert.Equal(t, len(ppppp.CAs), 1)

	pppppp, warn, err := NewCAPoolFromBytes(append([]byte(p256), []byte(v2)...))
	assert.Nil(t, err)
	assert.True(t, errors.Is(warn[0], ErrInvalidPEMCertificateUnsupported))
	assert.Equal(t, pppppp.CAs[string("a7938893ec8c4ef769b06d7f425e5e46f7a7f5ffa49c3bcf4a86b608caba9159")].Details.Name, rootCAP256.Details.Name)
	assert.Equal(t, len(pppppp.CAs), 1)
}

func appendByteSlices(b ...[]byte) []byte {
	retSlice := []byte{}
	for _, v := range b {
		retSlice = append(retSlice, v...)
	}
	return retSlice
}

func TestUnmrshalCertPEM(t *testing.T) {
	goodCert := []byte(`
# A good cert
-----BEGIN NEBULA CERTIFICATE-----
CkAKDm5lYnVsYSByb290IGNhKJfap9AFMJfg1+YGOiCUQGByMuNRhIlQBOyzXWbL
vcKBwDhov900phEfJ5DN3kABEkDCq5R8qBiu8sl54yVfgRcQXEDt3cHr8UTSLszv
bzBEr00kERQxxTzTsH8cpYEgRoipvmExvg8WP8NdAJEYJosB
-----END NEBULA CERTIFICATE-----
`)
	badBanner := []byte(`# A bad banner
-----BEGIN NOT A NEBULA CERTIFICATE-----
CkAKDm5lYnVsYSByb290IGNhKJfap9AFMJfg1+YGOiCUQGByMuNRhIlQBOyzXWbL
vcKBwDhov900phEfJ5DN3kABEkDCq5R8qBiu8sl54yVfgRcQXEDt3cHr8UTSLszv
bzBEr00kERQxxTzTsH8cpYEgRoipvmExvg8WP8NdAJEYJosB
-----END NOT A NEBULA CERTIFICATE-----
`)
	invalidPem := []byte(`# Not a valid PEM format
-BEGIN NEBULA CERTIFICATE-----
CkAKDm5lYnVsYSByb290IGNhKJfap9AFMJfg1+YGOiCUQGByMuNRhIlQBOyzXWbL
vcKBwDhov900phEfJ5DN3kABEkDCq5R8qBiu8sl54yVfgRcQXEDt3cHr8UTSLszv
bzBEr00kERQxxTzTsH8cpYEgRoipvmExvg8WP8NdAJEYJosB
-END NEBULA CERTIFICATE----`)

	certBundle := appendByteSlices(goodCert, badBanner, invalidPem)

	// Success test case
	cert, rest, err := UnmarshalNebulaCertificateFromPEM(certBundle)
	assert.NotNil(t, cert)
	assert.Equal(t, rest, append(badBanner, invalidPem...))
	assert.Nil(t, err)

	// Fail due to invalid banner.
	cert, rest, err = UnmarshalNebulaCertificateFromPEM(rest)
	assert.Nil(t, cert)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "bytes did not contain a proper nebula certificate banner")

	// Fail due to ivalid PEM format, because
	// it's missing the requisite pre-encapsulation boundary.
	cert, rest, err = UnmarshalNebulaCertificateFromPEM(rest)
	assert.Nil(t, cert)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "input did not contain a valid PEM encoded block")
}

func TestUnmarshalSigningPrivateKey(t *testing.T) {
	privKey := []byte(`# A good key
-----BEGIN NEBULA ED25519 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
-----END NEBULA ED25519 PRIVATE KEY-----
`)
	privP256Key := []byte(`# A good key
-----BEGIN NEBULA ECDSA P256 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NEBULA ECDSA P256 PRIVATE KEY-----
`)
	shortKey := []byte(`# A short key
-----BEGIN NEBULA ED25519 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
-----END NEBULA ED25519 PRIVATE KEY-----
`)
	invalidBanner := []byte(`# Invalid banner
-----BEGIN NOT A NEBULA PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
-----END NOT A NEBULA PRIVATE KEY-----
`)
	invalidPem := []byte(`# Not a valid PEM format
-BEGIN NEBULA ED25519 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
-END NEBULA ED25519 PRIVATE KEY-----`)

	keyBundle := appendByteSlices(privKey, privP256Key, shortKey, invalidBanner, invalidPem)

	// Success test case
	k, rest, curve, err := UnmarshalSigningPrivateKey(keyBundle)
	assert.Len(t, k, 64)
	assert.Equal(t, rest, appendByteSlices(privP256Key, shortKey, invalidBanner, invalidPem))
	assert.Equal(t, Curve_CURVE25519, curve)
	assert.Nil(t, err)

	// Success test case
	k, rest, curve, err = UnmarshalSigningPrivateKey(rest)
	assert.Len(t, k, 32)
	assert.Equal(t, rest, appendByteSlices(shortKey, invalidBanner, invalidPem))
	assert.Equal(t, Curve_P256, curve)
	assert.Nil(t, err)

	// Fail due to short key
	k, rest, curve, err = UnmarshalSigningPrivateKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, appendByteSlices(invalidBanner, invalidPem))
	assert.EqualError(t, err, "key was not 64 bytes, is invalid Ed25519 private key")

	// Fail due to invalid banner
	k, rest, curve, err = UnmarshalSigningPrivateKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "bytes did not contain a proper nebula Ed25519/ECDSA private key banner")

	// Fail due to ivalid PEM format, because
	// it's missing the requisite pre-encapsulation boundary.
	k, rest, curve, err = UnmarshalSigningPrivateKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "input did not contain a valid PEM encoded block")
}

func TestDecryptAndUnmarshalSigningPrivateKey(t *testing.T) {
	passphrase := []byte("DO NOT USE THIS KEY")
	privKey := []byte(`# A good key
-----BEGIN NEBULA ED25519 ENCRYPTED PRIVATE KEY-----
CjwKC0FFUy0yNTYtR0NNEi0IExCAgIABGAEgBCognnjujd67Vsv99p22wfAjQaDT
oCMW1mdjkU3gACKNW4MSXOWR9Sts4C81yk1RUku2gvGKs3TB9LYoklLsIizSYOLl
+Vs//O1T0I1Xbml2XBAROsb/VSoDln/6LMqR4B6fn6B3GOsLBBqRI8daDl9lRMPB
qrlJ69wer3ZUHFXA
-----END NEBULA ED25519 ENCRYPTED PRIVATE KEY-----
`)
	shortKey := []byte(`# A key which, once decrypted, is too short
-----BEGIN NEBULA ED25519 ENCRYPTED PRIVATE KEY-----
CjwKC0FFUy0yNTYtR0NNEi0IExCAgIABGAEgBCoga5h8owMEBWRSMMJKzuUvWce7
k0qlBkQmCxiuLh80MuASW70YcKt8jeEIS2axo2V6zAKA9TSMcCsJW1kDDXEtL/xe
GLF5T7sDl5COp4LU3pGxpV+KoeQ/S3gQCAAcnaOtnJQX+aSDnbO3jCHyP7U9CHbs
rQr3bdH3Oy/WiYU=
-----END NEBULA ED25519 ENCRYPTED PRIVATE KEY-----
`)
	invalidBanner := []byte(`# Invalid banner (not encrypted)
-----BEGIN NEBULA ED25519 PRIVATE KEY-----
bWRp2CTVFhW9HD/qCd28ltDgK3w8VXSeaEYczDWos8sMUBqDb9jP3+NYwcS4lURG
XgLvodMXZJuaFPssp+WwtA==
-----END NEBULA ED25519 PRIVATE KEY-----
`)
	invalidPem := []byte(`# Not a valid PEM format
-BEGIN NEBULA ED25519 ENCRYPTED PRIVATE KEY-----
CjwKC0FFUy0yNTYtR0NNEi0IExCAgIABGAEgBCognnjujd67Vsv99p22wfAjQaDT
oCMW1mdjkU3gACKNW4MSXOWR9Sts4C81yk1RUku2gvGKs3TB9LYoklLsIizSYOLl
+Vs//O1T0I1Xbml2XBAROsb/VSoDln/6LMqR4B6fn6B3GOsLBBqRI8daDl9lRMPB
qrlJ69wer3ZUHFXA
-END NEBULA ED25519 ENCRYPTED PRIVATE KEY-----
`)

	keyBundle := appendByteSlices(privKey, shortKey, invalidBanner, invalidPem)

	// Success test case
	curve, k, rest, err := DecryptAndUnmarshalSigningPrivateKey(passphrase, keyBundle)
	assert.Nil(t, err)
	assert.Equal(t, Curve_CURVE25519, curve)
	assert.Len(t, k, 64)
	assert.Equal(t, rest, appendByteSlices(shortKey, invalidBanner, invalidPem))

	// Fail due to short key
	curve, k, rest, err = DecryptAndUnmarshalSigningPrivateKey(passphrase, rest)
	assert.EqualError(t, err, "key was not 64 bytes, is invalid ed25519 private key")
	assert.Nil(t, k)
	assert.Equal(t, rest, appendByteSlices(invalidBanner, invalidPem))

	// Fail due to invalid banner
	curve, k, rest, err = DecryptAndUnmarshalSigningPrivateKey(passphrase, rest)
	assert.EqualError(t, err, "bytes did not contain a proper nebula encrypted Ed25519/ECDSA private key banner")
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)

	// Fail due to ivalid PEM format, because
	// it's missing the requisite pre-encapsulation boundary.
	curve, k, rest, err = DecryptAndUnmarshalSigningPrivateKey(passphrase, rest)
	assert.EqualError(t, err, "input did not contain a valid PEM encoded block")
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)

	// Fail due to invalid passphrase
	curve, k, rest, err = DecryptAndUnmarshalSigningPrivateKey([]byte("invalid passphrase"), privKey)
	assert.EqualError(t, err, "invalid passphrase or corrupt private key")
	assert.Nil(t, k)
	assert.Equal(t, rest, []byte{})
}

func TestEncryptAndMarshalSigningPrivateKey(t *testing.T) {
	// Having proved that decryption works correctly above, we can test the
	// encryption function produces a value which can be decrypted
	passphrase := []byte("passphrase")
	bytes := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	kdfParams := NewArgon2Parameters(64*1024, 4, 3)
	key, err := EncryptAndMarshalSigningPrivateKey(Curve_CURVE25519, bytes, passphrase, kdfParams)
	assert.Nil(t, err)

	// Verify the "key" can be decrypted successfully
	curve, k, rest, err := DecryptAndUnmarshalSigningPrivateKey(passphrase, key)
	assert.Len(t, k, 64)
	assert.Equal(t, Curve_CURVE25519, curve)
	assert.Equal(t, rest, []byte{})
	assert.Nil(t, err)

	// EncryptAndMarshalEd25519PrivateKey does not create any errors itself
}

func TestUnmarshalPrivateKey(t *testing.T) {
	privKey := []byte(`# A good key
-----BEGIN NEBULA X25519 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NEBULA X25519 PRIVATE KEY-----
`)
	privP256Key := []byte(`# A good key
-----BEGIN NEBULA P256 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NEBULA P256 PRIVATE KEY-----
`)
	shortKey := []byte(`# A short key
-----BEGIN NEBULA X25519 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
-----END NEBULA X25519 PRIVATE KEY-----
`)
	invalidBanner := []byte(`# Invalid banner
-----BEGIN NOT A NEBULA PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NOT A NEBULA PRIVATE KEY-----
`)
	invalidPem := []byte(`# Not a valid PEM format
-BEGIN NEBULA X25519 PRIVATE KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-END NEBULA X25519 PRIVATE KEY-----`)

	keyBundle := appendByteSlices(privKey, privP256Key, shortKey, invalidBanner, invalidPem)

	// Success test case
	k, rest, curve, err := UnmarshalPrivateKey(keyBundle)
	assert.Len(t, k, 32)
	assert.Equal(t, rest, appendByteSlices(privP256Key, shortKey, invalidBanner, invalidPem))
	assert.Equal(t, Curve_CURVE25519, curve)
	assert.Nil(t, err)

	// Success test case
	k, rest, curve, err = UnmarshalPrivateKey(rest)
	assert.Len(t, k, 32)
	assert.Equal(t, rest, appendByteSlices(shortKey, invalidBanner, invalidPem))
	assert.Equal(t, Curve_P256, curve)
	assert.Nil(t, err)

	// Fail due to short key
	k, rest, curve, err = UnmarshalPrivateKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, appendByteSlices(invalidBanner, invalidPem))
	assert.EqualError(t, err, "key was not 32 bytes, is invalid CURVE25519 private key")

	// Fail due to invalid banner
	k, rest, curve, err = UnmarshalPrivateKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "bytes did not contain a proper nebula private key banner")

	// Fail due to ivalid PEM format, because
	// it's missing the requisite pre-encapsulation boundary.
	k, rest, curve, err = UnmarshalPrivateKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "input did not contain a valid PEM encoded block")
}

func TestUnmarshalEd25519PublicKey(t *testing.T) {
	pubKey := []byte(`# A good key
-----BEGIN NEBULA ED25519 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NEBULA ED25519 PUBLIC KEY-----
`)
	shortKey := []byte(`# A short key
-----BEGIN NEBULA ED25519 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
-----END NEBULA ED25519 PUBLIC KEY-----
`)
	invalidBanner := []byte(`# Invalid banner
-----BEGIN NOT A NEBULA PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NOT A NEBULA PUBLIC KEY-----
`)
	invalidPem := []byte(`# Not a valid PEM format
-BEGIN NEBULA ED25519 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-END NEBULA ED25519 PUBLIC KEY-----`)

	keyBundle := appendByteSlices(pubKey, shortKey, invalidBanner, invalidPem)

	// Success test case
	k, rest, err := UnmarshalEd25519PublicKey(keyBundle)
	assert.Equal(t, len(k), 32)
	assert.Nil(t, err)
	assert.Equal(t, rest, appendByteSlices(shortKey, invalidBanner, invalidPem))

	// Fail due to short key
	k, rest, err = UnmarshalEd25519PublicKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, appendByteSlices(invalidBanner, invalidPem))
	assert.EqualError(t, err, "key was not 32 bytes, is invalid ed25519 public key")

	// Fail due to invalid banner
	k, rest, err = UnmarshalEd25519PublicKey(rest)
	assert.Nil(t, k)
	assert.EqualError(t, err, "bytes did not contain a proper nebula Ed25519 public key banner")
	assert.Equal(t, rest, invalidPem)

	// Fail due to ivalid PEM format, because
	// it's missing the requisite pre-encapsulation boundary.
	k, rest, err = UnmarshalEd25519PublicKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "input did not contain a valid PEM encoded block")
}

func TestUnmarshalX25519PublicKey(t *testing.T) {
	pubKey := []byte(`# A good key
-----BEGIN NEBULA X25519 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NEBULA X25519 PUBLIC KEY-----
`)
	pubP256Key := []byte(`# A good key
-----BEGIN NEBULA P256 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAA=
-----END NEBULA P256 PUBLIC KEY-----
`)
	shortKey := []byte(`# A short key
-----BEGIN NEBULA X25519 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==
-----END NEBULA X25519 PUBLIC KEY-----
`)
	invalidBanner := []byte(`# Invalid banner
-----BEGIN NOT A NEBULA PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END NOT A NEBULA PUBLIC KEY-----
`)
	invalidPem := []byte(`# Not a valid PEM format
-BEGIN NEBULA X25519 PUBLIC KEY-----
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-END NEBULA X25519 PUBLIC KEY-----`)

	keyBundle := appendByteSlices(pubKey, pubP256Key, shortKey, invalidBanner, invalidPem)

	// Success test case
	k, rest, curve, err := UnmarshalPublicKey(keyBundle)
	assert.Equal(t, len(k), 32)
	assert.Nil(t, err)
	assert.Equal(t, rest, appendByteSlices(pubP256Key, shortKey, invalidBanner, invalidPem))
	assert.Equal(t, Curve_CURVE25519, curve)

	// Success test case
	k, rest, curve, err = UnmarshalPublicKey(rest)
	assert.Equal(t, len(k), 65)
	assert.Nil(t, err)
	assert.Equal(t, rest, appendByteSlices(shortKey, invalidBanner, invalidPem))
	assert.Equal(t, Curve_P256, curve)

	// Fail due to short key
	k, rest, curve, err = UnmarshalPublicKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, appendByteSlices(invalidBanner, invalidPem))
	assert.EqualError(t, err, "key was not 32 bytes, is invalid CURVE25519 public key")

	// Fail due to invalid banner
	k, rest, curve, err = UnmarshalPublicKey(rest)
	assert.Nil(t, k)
	assert.EqualError(t, err, "bytes did not contain a proper nebula public key banner")
	assert.Equal(t, rest, invalidPem)

	// Fail due to ivalid PEM format, because
	// it's missing the requisite pre-encapsulation boundary.
	k, rest, curve, err = UnmarshalPublicKey(rest)
	assert.Nil(t, k)
	assert.Equal(t, rest, invalidPem)
	assert.EqualError(t, err, "input did not contain a valid PEM encoded block")
}

// Ensure that upgrading the protobuf library does not change how certificates
// are marshalled, since this would break signature verification
func TestMarshalingNebulaCertificateConsistency(t *testing.T) {
	before := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	after := time.Date(2017, time.January, 18, 28, 40, 0, 0, time.UTC)
	pubKey := []byte("1234567890abcedfghij1234567890ab")

	nc := NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name: "testing",
			Ips: []*net.IPNet{
				{IP: net.ParseIP("10.1.1.1"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("10.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
				{IP: net.ParseIP("10.1.1.3"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
			},
			Subnets: []*net.IPNet{
				{IP: net.ParseIP("9.1.1.1"), Mask: net.IPMask(net.ParseIP("255.0.255.0"))},
				{IP: net.ParseIP("9.1.1.2"), Mask: net.IPMask(net.ParseIP("255.255.255.0"))},
				{IP: net.ParseIP("9.1.1.3"), Mask: net.IPMask(net.ParseIP("255.255.0.0"))},
			},
			Groups:    []string{"test-group1", "test-group2", "test-group3"},
			NotBefore: before,
			NotAfter:  after,
			PublicKey: pubKey,
			IsCA:      false,
			Issuer:    "1234567890abcedfghij1234567890ab",
		},
		Signature: []byte("1234567890abcedfghij1234567890ab"),
	}

	b, err := nc.Marshal()
	assert.Nil(t, err)
	//t.Log("Cert size:", len(b))
	assert.Equal(t, "0aa2010a0774657374696e67121b8182845080feffff0f828284508080fcff0f8382845080fe83f80f1a1b8182844880fe83f80f8282844880feffff0f838284488080fcff0f220b746573742d67726f757031220b746573742d67726f757032220b746573742d67726f75703328f0e0e7d70430a08681c4053a20313233343536373839306162636564666768696a3132333435363738393061624a081234567890abcedf1220313233343536373839306162636564666768696a313233343536373839306162", fmt.Sprintf("%x", b))

	b, err = proto.Marshal(nc.getRawDetails())
	assert.Nil(t, err)
	//t.Log("Raw cert size:", len(b))
	assert.Equal(t, "0a0774657374696e67121b8182845080feffff0f828284508080fcff0f8382845080fe83f80f1a1b8182844880fe83f80f8282844880feffff0f838284488080fcff0f220b746573742d67726f757031220b746573742d67726f757032220b746573742d67726f75703328f0e0e7d70430a08681c4053a20313233343536373839306162636564666768696a3132333435363738393061624a081234567890abcedf", fmt.Sprintf("%x", b))
}

func TestNebulaCertificate_Copy(t *testing.T) {
	ca, _, caKey, err := newTestCaCert(time.Now(), time.Now().Add(10*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)

	c, _, _, err := newTestCert(ca, caKey, time.Now(), time.Now().Add(5*time.Minute), []*net.IPNet{}, []*net.IPNet{}, []string{})
	assert.Nil(t, err)
	cc := c.Copy()

	test.AssertDeepCopyEqual(t, c, cc)
}

func TestUnmarshalNebulaCertificate(t *testing.T) {
	// Test that we don't panic with an invalid certificate (#332)
	data := []byte("\x98\x00\x00")
	_, err := UnmarshalNebulaCertificate(data)
	assert.EqualError(t, err, "encoded Details was nil")
}

func newTestCaCert(before, after time.Time, ips, subnets []*net.IPNet, groups []string) (*NebulaCertificate, []byte, []byte, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if before.IsZero() {
		before = time.Now().Add(time.Second * -60).Round(time.Second)
	}
	if after.IsZero() {
		after = time.Now().Add(time.Second * 60).Round(time.Second)
	}

	nc := &NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name:           "test ca",
			NotBefore:      time.Unix(before.Unix(), 0),
			NotAfter:       time.Unix(after.Unix(), 0),
			PublicKey:      pub,
			IsCA:           true,
			InvertedGroups: make(map[string]struct{}),
		},
	}

	if len(ips) > 0 {
		nc.Details.Ips = ips
	}

	if len(subnets) > 0 {
		nc.Details.Subnets = subnets
	}

	if len(groups) > 0 {
		nc.Details.Groups = groups
	}

	err = nc.Sign(Curve_CURVE25519, priv)
	if err != nil {
		return nil, nil, nil, err
	}
	return nc, pub, priv, nil
}

func newTestCaCertP256(before, after time.Time, ips, subnets []*net.IPNet, groups []string) (*NebulaCertificate, []byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pub := elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y)
	rawPriv := priv.D.FillBytes(make([]byte, 32))

	if before.IsZero() {
		before = time.Now().Add(time.Second * -60).Round(time.Second)
	}
	if after.IsZero() {
		after = time.Now().Add(time.Second * 60).Round(time.Second)
	}

	nc := &NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name:           "test ca",
			NotBefore:      time.Unix(before.Unix(), 0),
			NotAfter:       time.Unix(after.Unix(), 0),
			PublicKey:      pub,
			IsCA:           true,
			Curve:          Curve_P256,
			InvertedGroups: make(map[string]struct{}),
		},
	}

	if len(ips) > 0 {
		nc.Details.Ips = ips
	}

	if len(subnets) > 0 {
		nc.Details.Subnets = subnets
	}

	if len(groups) > 0 {
		nc.Details.Groups = groups
	}

	err = nc.Sign(Curve_P256, rawPriv)
	if err != nil {
		return nil, nil, nil, err
	}
	return nc, pub, rawPriv, nil
}

func newTestCert(ca *NebulaCertificate, key []byte, before, after time.Time, ips, subnets []*net.IPNet, groups []string) (*NebulaCertificate, []byte, []byte, error) {
	issuer, err := ca.Sha256Sum()
	if err != nil {
		return nil, nil, nil, err
	}

	if before.IsZero() {
		before = time.Now().Add(time.Second * -60).Round(time.Second)
	}
	if after.IsZero() {
		after = time.Now().Add(time.Second * 60).Round(time.Second)
	}

	if len(groups) == 0 {
		groups = []string{"test-group1", "test-group2", "test-group3"}
	}

	if len(ips) == 0 {
		ips = []*net.IPNet{
			{IP: net.ParseIP("10.1.1.1").To4(), Mask: net.IPMask(net.ParseIP("255.255.255.0").To4())},
			{IP: net.ParseIP("10.1.1.2").To4(), Mask: net.IPMask(net.ParseIP("255.255.0.0").To4())},
			{IP: net.ParseIP("10.1.1.3").To4(), Mask: net.IPMask(net.ParseIP("255.0.255.0").To4())},
		}
	}

	if len(subnets) == 0 {
		subnets = []*net.IPNet{
			{IP: net.ParseIP("9.1.1.1").To4(), Mask: net.IPMask(net.ParseIP("255.0.255.0").To4())},
			{IP: net.ParseIP("9.1.1.2").To4(), Mask: net.IPMask(net.ParseIP("255.255.255.0").To4())},
			{IP: net.ParseIP("9.1.1.3").To4(), Mask: net.IPMask(net.ParseIP("255.255.0.0").To4())},
		}
	}

	var pub, rawPriv []byte

	switch ca.Details.Curve {
	case Curve_CURVE25519:
		pub, rawPriv = x25519Keypair()
	case Curve_P256:
		pub, rawPriv = p256Keypair()
	default:
		return nil, nil, nil, fmt.Errorf("unknown curve: %v", ca.Details.Curve)
	}

	nc := &NebulaCertificate{
		Details: NebulaCertificateDetails{
			Name:           "testing",
			Ips:            ips,
			Subnets:        subnets,
			Groups:         groups,
			NotBefore:      time.Unix(before.Unix(), 0),
			NotAfter:       time.Unix(after.Unix(), 0),
			PublicKey:      pub,
			IsCA:           false,
			Curve:          ca.Details.Curve,
			Issuer:         issuer,
			InvertedGroups: make(map[string]struct{}),
		},
	}

	err = nc.Sign(ca.Details.Curve, key)
	if err != nil {
		return nil, nil, nil, err
	}

	return nc, pub, rawPriv, nil
}

func x25519Keypair() ([]byte, []byte) {
	privkey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privkey); err != nil {
		panic(err)
	}

	pubkey, err := curve25519.X25519(privkey, curve25519.Basepoint)
	if err != nil {
		panic(err)
	}

	return pubkey, privkey
}

func p256Keypair() ([]byte, []byte) {
	privkey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pubkey := privkey.PublicKey()
	return pubkey.Bytes(), privkey.Bytes()
}
