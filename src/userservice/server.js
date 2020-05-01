const path = require('path');
const grpc = require('grpc');
const protoLoader = require('@grpc/proto-loader');
const pino = require('pino');

const MAIN_PROTO_PATH = path.join(__dirname, './proto/app.proto');

const PORT = process.env.PORT;

const shopProto = _loadProto(MAIN_PROTO_PATH).shop;

const logger = pino({
    name: 'userservice-server',
    messageKey: 'message',
    changeLevelName: 'severity',
    useLevelLabels: true
})

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

/**
 *
 * @param call {UserId}
 * @param callback {err, User}
 */
function getUser(call, callback) {
    const request = call.request;

    try {
        const data = require('./data/users_db.json');

        for (let d of data) {
            if (d.id === request.value) {
                callback(null, d);
            }
        }
    } catch (e) {
        logger.error(`getting user ${request.value} failed: ${e}`);
        callback(e);
    }

}
/**
 * starts an RPC server that receives requests for the
 * User service at the sample server port
 */
function main() {
    logger.info(`starting gRPC server on port ${PORT}...`);
    const server = new grpc.Server();
    server.addService(shopProto.UserService.service, {getUser});
    server.bind(`0.0.0.0:${PORT}`, grpc.ServerCredentials.createInsecure());
    server.start();
}

main();