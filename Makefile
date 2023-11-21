deploy:
	gcloud functions deploy discorss --gen2 --runtime go121 --trigger-http --entry-point SendFeed --region=us-west1 --service-account discorss-srv-account@discorss.iam.gserviceaccount.com
