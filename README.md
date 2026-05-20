# Terraform Provider DNSHE

Terraform provider for [dnshe free domain](https://www.dnshe.com).

[doc](https://registry.terraform.io/providers/serialt/dnshe/latest/docs)

## Development

```bash
go tidy # install dependencies
make build # build the provider
make install # install the provider
```

Please see `./examples` for example usage.

## Testing

Then run the acceptance tests:

```bash
make testacc
```

## Generate docs

```bash
make generate
```

## References

https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework