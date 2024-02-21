IMAGE_NAME ?= arrudaricardo/rinha-de-backend-2024-q1-golang

release: 
	docker build --platform linux/amd64 --tag $(IMAGE_NAME):latest .
	docker push $(IMAGE_NAME):latest 

