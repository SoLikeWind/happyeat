.PHONY: api
# generate api service code
api:
	goctl api go --api app/api/v1/menu.api --dir app