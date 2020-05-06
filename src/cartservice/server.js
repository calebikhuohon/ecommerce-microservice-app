const path = require('path');
const grpc = require('grpc');
const protoLoader = require('@grpc/proto-loader');
const pino = require('pino');

const MAIN_PROTO_PATH = path.join(__dirname, './proto/app.proto');
const PORT = process.env.PORT || "5150";

/**
 * loads a protobuf file
 */

const _loadProto = path => {
    const packageDefinition = protoLoader.loadSync(
        path, {
            keepCase: true,
            longs: String,
            enums: String,
            defaults: true,
            oneofs: true
        }
    );

    return grpc.loadPackageDefinition(packageDefinition);
}

const shopProto = _loadProto(MAIN_PROTO_PATH).shop;

const logger = pino({
    name: 'cartservice-server',
    messageKey: 'message',
    changeLevelName: 'severity',
    useLevelLabels: true
});


/**
 * starts an RPC server that receives requests for the
 * User service at the sample server port
 */
function main() {
    logger.info(`starting gRPC server on port ${PORT}...`);
    const server = new grpc.Server();
    server.addService(shopProto.CartService.service, {});
    server.bind(`0.0.0.0:${PORT}`, grpc.ServerCredentials.createInsecure());
    server.start();
}

main();
