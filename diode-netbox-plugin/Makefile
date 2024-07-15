.PHONY: docker-compose-netbox-plugin-up docker-compose-netbox-plugin-down docker-compose-netbox-plugin-test

ifneq ($(shell docker compose version 2>/dev/null),)
  DOCKER_COMPOSE := docker compose
else
  DOCKER_COMPOSE := docker-compose
endif

docker-compose-netbox-plugin-up:
	@$(DOCKER_COMPOSE) -f ../diode-server/docker/docker-compose.netbox.yaml up -d --build

docker-compose-netbox-plugin-down:
	@$(DOCKER_COMPOSE) -f ../diode-server/docker/docker-compose.netbox.yaml down

docker-compose-netbox-plugin-test:
	-@$(DOCKER_COMPOSE) -f ../diode-server/docker/docker-compose.netbox.yaml run --rm netbox ./manage.py test --keepdb netbox_diode_plugin
	$(MAKE) docker-compose-netbox-plugin-down
