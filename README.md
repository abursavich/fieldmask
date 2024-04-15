# Fieldmask
[![License][license-img]][license]
[![GoDev Reference][godev-img]][godev]
[![Go Report Card][goreportcard-img]][goreportcard]

Package fieldmask provides protobuf field masking for projections and updates.

It follows [Google API Improvement Proposal] guidelines by default, but provides options that
allow it to fall back to the behavior of basic [FieldMask] documentation.



[license]: https://raw.githubusercontent.com/abursavich/fieldmask/main/LICENSE
[license-img]: https://img.shields.io/badge/license-mit-blue.svg?style=for-the-badge

[godev]: https://pkg.go.dev/bursavich.dev/fieldmask
[godev-img]: https://img.shields.io/static/v1?logo=go&logoColor=white&color=00ADD8&label=dev&message=reference&style=for-the-badge

[goreportcard]: https://goreportcard.com/report/bursavich.dev/fieldmask
[goreportcard-img]: https://goreportcard.com/badge/bursavich.dev/fieldmask?style=for-the-badge

[Google API Improvement Proposal]: https://google.aip.dev/
[FieldMask]: https://protobuf.dev/reference/protobuf/google.protobuf/#field-mask
