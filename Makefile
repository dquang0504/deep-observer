COMPOSE_DIR := deploy/compose
OBSERVER_COMPOSE := $(COMPOSE_DIR)/docker-compose.observer.yml

.PHONY: observer-up observer-down observer-logs observer-reset

observer-run:
	cd $(COMPOSE_DIR) && docker compose -f docker-compose.observer.yml up -d

observer-up:
	cd $(COMPOSE_DIR) && docker compose -f docker-compose.observer.yml up -d --build

observer-down:
	cd $(COMPOSE_DIR) && docker compose -f docker-compose.observer.yml down

observer-logs:
	cd $(COMPOSE_DIR) && docker compose -f docker-compose.observer.yml logs -f --tail=200

observer-reset:
	cd $(COMPOSE_DIR) && docker compose -f docker-compose.observer.yml down -v
	cd $(COMPOSE_DIR) && docker compose -f docker-compose.observer.yml up -d --build