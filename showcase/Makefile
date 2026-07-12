.PHONY: generate css build dev clean browser browser-stop smoke

# Must match the playwright-go minor version in go.mod (v0.6000.x -> 1.60.x).
PLAYWRIGHT_VERSION = 1.60.0

browser:
	docker run -d --rm --name loom-playwright -p 43110:43110 \
		--add-host=host.docker.internal:host-gateway \
		mcr.microsoft.com/playwright:v$(PLAYWRIGHT_VERSION)-noble \
		npx -y playwright@$(PLAYWRIGHT_VERSION) run-server --port 43110 --host 0.0.0.0

browser-stop:
	docker stop loom-playwright

smoke:
	go test -v -run 'Test' -count=1 .

generate:
	templ generate

css:
	go run github.com/pietjan/loom/cmd/css -o css/input.css
	tailwindcss -i css/input.css -o static/styles.css --minify

build: generate css
	go build .

dev: generate css
	go run .

clean:
	rm -f css/input.css static/styles.css templates/*_templ.go
