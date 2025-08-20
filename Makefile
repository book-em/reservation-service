.PHONY: run

# Usage:
# make [command] {MODE}
#
# command := run (default) / test / test_unit / test_integration
# MODE :=    ci / local
#

run:
	echo "Run using infrastructure/"

test: test_integration

test_unit:
	./run-tests.sh

test_integration:
	./run-integration.sh $(MODE)