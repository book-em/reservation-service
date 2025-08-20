# Book'em reservation microservice

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![Bash Script](https://img.shields.io/badge/bash_script-%23121011.svg?style=for-the-badge&logo=gnu-bash&logoColor=white)

## Getting started

Use `make`. The app is run as part of [book-em/infrastructure](https://github.com/book-em/infrastructure),
while the tests are either run locally (unit) or through docker compose (integration). 

## Contributing guidelines

1) Follow [Feature Branch Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/feature-branch-workflow)
2) Use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/):
    - `feat!:` changes major version
    - `feat:` changes minor version
    - `fix:` changes patch version
    - Don't change major version until the service is useable

# License

This project uses the BSD 2-Clause License. See `LICENSE` for more info.