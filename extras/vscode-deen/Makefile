IMAGE ?= node:22-alpine
NAME ?= vscode-deen
RUN = docker exec $(NAME)

package: start run-package-inside stop

update: start run-update-inside stop

start:
	docker run -dit --platform linux/amd64 -v $(shell pwd)\:/app --name=$(NAME) --rm --init $(IMAGE) /bin/sh

run-package-inside:
	$(RUN) /bin/sh -c "cd /app && npm install"
	$(RUN) /bin/sh -c "cd /app && npm run compile"
	$(RUN) /bin/sh -c "cd /app && npm exec -- @vscode/vsce package"

run-update-inside:
	$(RUN) /bin/sh -c "cd /app && npm audit"
	$(RUN) /bin/sh -c "cd /app && npm audit fix"

stop:
	docker stop $(NAME)

.PHONY: package update start run-package-inside run-update-inside stop
