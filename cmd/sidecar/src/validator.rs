use tonic::{Request, Response, Status};
use ring::digest::{Context, SHA256};
use minio_rsc::client::{Minio, MinioBuilder};
use minio_rsc::error::Result as MinioResult;
use minio_rsc::provider::StaticProvider;

pub mod drtest {
    tonic::include_proto!("drtest");
}

use drtest::{DataRequest, DataResponse};

pub struct Validator {
    minio: Minio,
}

impl Validator {
    pub fn new() -> Self {
        let endpoint = "http://minio:9000";
        let access_key = std::env::var("MINIO_ACCESS_KEY").unwrap_or_default();
        let secret_key = std::env::var("MINIO_SECRET_KEY").unwrap_or_default();
        
        let provider = StaticProvider::new(&access_key, &secret_key, None); // Create StaticProvider
        let minio = MinioBuilder::new()
            .endpoint(endpoint)
            .secure(false) // Explicitly disable TLS for HTTP endpoint
            .provider(provider) // Set credentials provider
            .build()
            .expect("Failed to init MinIO client");
            
        Validator { minio }
    }

    fn compute_checksum(data: &[u8]) -> String {
        let mut context = Context::new(&SHA256);
        context.update(data);
        hex::encode(context.finish().as_ref())
    }

    async fn upload_to_minio(&self, bucket: &str, object: &str, data: Vec<u8>) -> MinioResult<()> {
        self.minio.put_object(bucket, object, data.into()).await?;
        Ok(())
    }
}

#[tonic::async_trait]
impl drtest::data_validator_server::DataValidator for Validator {
    async fn validate_data(
        &self,
        request: Request<DataRequest>,
    ) -> Result<Response<DataResponse>, Status> {
        let req = request.into_inner();
        let data = req.data;
        let checksum = Self::compute_checksum(&data);
        
        if let Some(expected) = req.expected_checksum {
            if expected != checksum {
                return Ok(Response::new(DataResponse {
                    success: false,
                    checksum,
                    object_path: "".to_string(),
                    validation_error: format!("Checksum mismatch: expected {}, got {}", expected, checksum),
                }));
            }
        }

        let object = format!("validation-{}.bin", checksum);
        let data_vec = data.to_vec();
        
        self.upload_to_minio(&req.bucket, &object, data_vec)
            .await
            .map_err(|e| Status::internal(format!("MinIO upload failed: {}", e)))?;

        Ok(Response::new(DataResponse {
            success: true,
            checksum,
            object_path: format!("{}/{}", req.bucket, object),
            validation_error: "".to_string(),
        }))
    }
}
