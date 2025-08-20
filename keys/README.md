The public key in this repo is used for validating JWTs.

This service doesn't create JWTs, so it doesn't need the private key.

For ease of use during development, the public key is in `keys.tar.gz`. Please
use this one.

The key is mounted in `compose.yml` files across this and other services.
