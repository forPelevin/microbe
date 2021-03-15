up:
	docker-compose -f docker-compose.yml up -d --force-recreate

up_env:
	docker-compose -f docker-compose.yml up -d redis mongo

down:
	docker-compose -f docker-compose.yml down