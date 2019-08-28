.PHONY:  default  diag  refresh  run  test  test-coverage  test-docker  test-release

default: test

diag:
	(cd scripts/seqdiag ; go run main.go ../../docs/*.seq)

refresh:
	cookiecutter gh:sjansen/cookiecutter-golang --output-dir .. --config-file .cookiecutter.yaml --no-input --overwrite-if-exists
	git checkout go.mod go.sum

run:
	docker-compose \
		-f docker-compose.yml \
		-f docker-compose.override.yml \
		up dynamodb

test:
	@scripts/run-all-tests
	@echo ========================================
	@git grep TODO  -- '**.go' || true
	@git grep FIXME -- '**.go' || true

test-coverage: test-docker
	go tool cover -html=dist/coverage.txt

test-docker:
	@scripts/docker-up-test

test-release:
	git stash -u -k
	goreleaser release --rm-dist --skip-publish
	-git stash pop
