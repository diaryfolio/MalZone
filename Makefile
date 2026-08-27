.PHONY: check design-check

check: design-check

design-check:
	python3 -m unittest discover -s tests -p 'test_*.py'

