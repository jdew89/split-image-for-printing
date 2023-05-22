ifneq (,)
.error This Makefile requires GNU Make.
endif

ifneq ("$(wildcard ./.env)","")
    include .env
	export
endif

sam_template_base=sam
sam_template=$(sam_template_base).yaml
sam_output=$(sam_template_base)-package.yaml

AWS_REGION       	?=us-east-2
AWS_PROFILE		 	?=ee

STAGE            	?=prod

FUNCTION_NAME_PREFIX ?=split-images-to-pdf-pages
FUNCTION_NAME		?=${FUNCTION_NAME_PREFIX}

# Bucket where the SAM deloyment files are kept

S3_BUCKET 			?=cloud.formation.${STAGE}

# can NOT contain periods, but can use hyphens
STACK_NAME       	?=${FUNCTION_NAME_PREFIX}

buildgo: 
	rm dist/main.zip & GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/main && zip -j dist/main.zip dist/main dist/index.html

build: 
	sam build \
	    --template-file $(sam_template) \
		--region ${AWS_REGION}
# --use-container

# Building and deploying app
package: buildgo
	sam package \
	    --s3-bucket $(S3_BUCKET) \
		--s3-prefix ${FUNCTION_NAME_PREFIX} \
	    --template-file $(sam_template) \
	    --output-template-file $(sam_output) \
		--region ${AWS_REGION} \
		--profile ${AWS_PROFILE}

deploy: buildgo
	aws --profile ${AWS_PROFILE} lambda update-function-code --function-name ${FUNCTION_NAME} --zip-file fileb://dist/main.zip --no-paginate --output table


# Doesnt build and delpoy image
deployold: 
	sam deploy \
		--profile ${AWS_PROFILE} 

destroy: 
	sam delete --stack-name $(STACK_NAME)

forcedestroy:
	sam delete --stack-name $(STACK_NAME) --region ${AWS_REGION} --no-prompts

local:
	sam local start-api -t sam.yaml