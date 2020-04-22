if [ "$#" -eq 1 ]; then
    go build main_firebase.go
    enterprise=$1
    id='smart-quasar-274315'
    echo "gcr.io/$id/$enterprise"
    gcloud config set project $id
    gcloud config set run/region europe-west1
    gcloud builds submit --tag gcr.io/$id/$enterprise
    gcloud run deploy --image gcr.io/$id/$enterprise --platform managed --set-env-vars ENTERPRISE=$enterprise
else
    echo "invalid number of arguments"
fi