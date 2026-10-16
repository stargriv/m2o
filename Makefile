VENV := .venv
PYTHON := $(VENV)/bin/python
PYTHON_DIR := python

.DEFAULT_GOAL := run
.PHONY: run setup test go-build go-run go-test

# The default target deliberately does not create the environment.
run:
	@if [ ! -d "$(VENV)" ]; then \
		echo "$(VENV) does not exist. Run 'make setup' first."; \
		exit 1; \
	fi
	@cd $(PYTHON_DIR) && ../$(PYTHON) main.py

# Create and populate the virtual environment only when it is absent.
setup:
	@if [ -d "$(VENV)" ]; then \
		echo "$(VENV) already exists; nothing to do."; \
	else \
		python3 -m venv "$(VENV)" && \
		$(PYTHON) -m pip install --upgrade pip && \
		$(PYTHON) -m pip install -r $(PYTHON_DIR)/requirements.txt; \
	fi

test:
	@if [ ! -d "$(VENV)" ]; then \
		echo "$(VENV) does not exist. Run 'make setup' first."; \
		exit 1; \
	fi
	@cd $(PYTHON_DIR) && ../$(PYTHON) -m unittest discover -s tests

# The Go player uses only the standard library and produces one macOS binary.
go-build:
	@mkdir -p bin
	@cd go && go build -trimpath -ldflags='-s -w' -o ../bin/m2o ./cmd/m2o

go-run: go-build
	@./bin/m2o

go-test:
	@cd go && go test ./...
