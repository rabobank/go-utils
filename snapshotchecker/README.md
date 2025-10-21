### snapshotchecker

`snapshotchecker` queries AWS for RDS databases, for each database it checks if the most recent snapshot is older than a specified threshold. 
If it is, it prepares an email to a contact address.
This contact address is found by getting the org and space name from the tags of the RDS instance, the looking in the orgs-spaces repo to find the contact email for that org and space.
It outputs messages and errors to stderr, and the mails to stdout in a json format understandable by the concourse mail resource.

## environment variables
* THRESHOLD_DAYS - number of days after which a snapshot is considered old
* AWS_REGION - AWS region to query RDS instances in
* ORGS_SPACES_REPO_PATH - path to the cloned orgs-spaces repo
* MAIL_CC - email address to CC on all emails
