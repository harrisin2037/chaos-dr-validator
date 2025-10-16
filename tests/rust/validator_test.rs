  #[cfg(test)]
  mod tests {
      use super::*;
      use tokio::runtime::Runtime;

      #[test]
      fn test_compute_checksum() {
          let data = b"test-data";
          let checksum = Validator::compute_checksum(data);
          assert_eq!(checksum, "d5579c46dfcc7f18207013e65b44e4cb4e2c2298f4ac457ba8f82743f31e930b");
      }

      #[tokio::test]
      async fn test_upload_to_minio() {
          let validator = Validator::new();
          let result = validator.upload_to_minio("test-bucket", "test-object", b"test-data".to_vec()).await;
          assert!(result.is_ok());
      }
  }