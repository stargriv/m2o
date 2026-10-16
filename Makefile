VENV := .venv
PYTHON := $(VENV)/bin/python

.DEFAULT_GOAL := run
.PHONY: run setup test

# The default target deliberately does not create the environment.
run:
	@if [ ! -d "$(VENV)" ]; then \
		echo "$(VENV) does not exist. Run 'make setup' first."; \
		exit 1; \
	fi
	@$(PYTHON) main.py

# Create and populate the virtual environment only when it is absent.
setup:
	@if [ -d "$(VENV)" ]; then \
		echo "$(VENV) already exists; nothing to do."; \
	else \
		python3 -m venv "$(VENV)" && \
		$(PYTHON) -m pip install --upgrade pip && \
		$(PYTHON) -m pip install -r requirements.txt; \
	fi

test:
	@if [ ! -d "$(VENV)" ]; then \
		echo "$(VENV) does not exist. Run 'make setup' first."; \
		exit 1; \
	fi
	@$(PYTHON) -m unittest discover -s tests
