# FEATURE

I want to have a Go API that runs aa a AWS Lambda function. The objects that are managed by the API are songs and
comments for a song. A song has the following attributes:

- ID
- Title
- Subtitle
- Description
- image urls
- audio urls
- lyrics
- stage
- tags

Also for each song there can be one or more comments. Each comment is associated with a certain audio file.

The data should be stored in DynamoDB. All necessary infrastructure should be created using terraform, so create a
separate folder with terraform modules.

## EXAMPLES & DOCUMENTATION

All examples are already referenced in prp_mcp_base.md - do any additional research as needed.

## OTHER CONSIDERATIONS

- Do not use complex regex or complex parsing patterns, we use an LLM to parse PRPs.
- Model and API key for Anthropic both need to be environment variables - these are set up in .dev.vars.example
- It's very important that we create one task per file to keep concerns separate