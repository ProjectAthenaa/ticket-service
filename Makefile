build:
	docker build --build-arg GH_TOKEN=$(token)  -t registry.digitalocean.com/athenabot/antibots/ticket:latest .