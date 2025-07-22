// Package csr implements helper functionality around the Confluent Schema Registry.
package csr

// Config is all the configuration needed to connect to a CSR Instance.
type Config struct {
	// URL of the CSR instance.
	//
	// The absence of this field says to not connect to the CSR.
	URL string
	// Username to use for authentication, if any.
	Username string
	// Password to use for authentication, if any.
	Password string
}
