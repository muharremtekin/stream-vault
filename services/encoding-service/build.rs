fn main() -> Result<(), Box<dyn std::error::Error>> {
    let proto_dir = "proto";

    if !std::path::Path::new(proto_dir).exists() {
        println!("cargo:warning=proto directory not found, skipping proto generation");
        println!("cargo:warning=Run scripts/generate-proto.sh to copy proto files");
        return Ok(());
    }

    let encoding_proto = "proto/encoding/v1/encoding.proto";
    if std::path::Path::new(encoding_proto).exists() {
        tonic_build::configure()
            .build_server(true)
            .build_client(false)
            .compile_protos(&[encoding_proto], &[proto_dir])?;

        println!("cargo:rerun-if-changed={}", encoding_proto);
    }

    println!("cargo:rerun-if-changed=proto/");
    Ok(())
}
