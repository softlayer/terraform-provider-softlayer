# `softlayer_security_certificate`

Create, update, and destroy [SoftLayer Security Certificates](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Security_Certificate).

**Using certs on file:**

```hcl
resource "softlayer_security_certificate" "test_cert" {
  certificate = "${file("cert.pem")}"
  private_key = "${file("key.pem")}"
}
```

**Example with cert in-line:**

```hcl
resource "softlayer_security_certificate" "test_cert" {
    certificate = <<EOF
[......] # cert contents
-----END CERTIFICATE-----
    EOF

    private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
[......] # cert contents
-----END RSA PRIVATE KEY-----
    EOF
}
```

## Argument Reference

* `certificate` | *string*
    * (Required) The certificate provided publicly to clients requesting identity credentials.
* `intermediate_certificate` | *string*
    * (Optional) The intermediate certificate authorities certificate that completes the certificate chain for the issued certificate. Required when clients will only trust the root certificate.
* `private_key` | *string*
    * (Required) The private key in the key/certificate pair.

## Attributes Reference

* `common_name` - The common name (usually a domain name) encoded within the certificate.
* `create_date` - The date the certificate record was created.
* `id` - The ID of the certificate record.
* `key_size` - The size (number of bits) of the public key represented by the certificate.
* `modify_date` - The date the certificate record was last modified.
* `organization_name` - The organizational name encoded in the certificate.
* `validity_begin` - The UTC timestamp representing the beginning of the certificate's validity.
* `validity_days` - The number of days remaining in the validity period for the certificate.
* `validity_end` - The UTC timestamp representing the end of the certificate's validity period.
