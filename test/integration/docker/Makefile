VAULT_CONTAINER := vault-new
DATA_DIR := ./data

.PHONY: restart clean stop

## Restart Vault (keeps data)
restart:
	@docker stop $(VAULT_CONTAINER) > /dev/null 2>&1 || true
	@docker rm $(VAULT_CONTAINER) > /dev/null 2>&1 || true
	@docker compose up -d

## Clean Vault (removes data directory)
clean:
	@docker stop $(VAULT_CONTAINER) > /dev/null 2>&1 || true
	@docker rm $(VAULT_CONTAINER) > /dev/null 2>&1 || true
	@rm -rf $(DATA_DIR)
	@docker compose up -d

## Stop Vault only
stop:
	@docker stop $(VAULT_CONTAINER) > /dev/null 2>&1 || true