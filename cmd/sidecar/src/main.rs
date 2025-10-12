use tonic::transport::Server;
use validator::drtest::data_validator_server::DataValidatorServer;

mod validator;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "0.0.0.0:50051".parse()?;
    let validator = validator::Validator::new();

    println!("Starting gRPC server on {}", addr);
    Server::builder()
        .add_service(DataValidatorServer::new(validator))
        .serve(addr)
        .await?;

    Ok(())
}
