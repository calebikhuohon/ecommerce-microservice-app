const path = require('path');
const grpc = require('grpc');
const protoLoader = require('@grpc/proto-loader');
const pino = require('pino');

const MAIN_PROTO_PATH = path.join(__dirname, './proto/app.proto');

const PORT = process.env.PORT;

const users = require('./users');

const logger = pino({
    name: 'userservice-server',
    messageKey: 'message',
    changeLevelName: 'severity',
    useLevelLabels: true
})

/**
 * loads a protobuf file
 */

class ShopServer {
    constructor(protoRoot, port = PORT) {
        this.port = port;

        this.packages = {
            shop: loadProto(MAIN_PROTO_PATH),
        };

        this.server = new grpc.Server();
        this.loadAllProtos(protoRoot);

    }

    listen() {
        this.server.bind(`0.0.0.0:${this.port}`, grpc.ServerCredentials.createInsecure());
        logger.info(`user service grpc server listening on ${this.port}`);
        this.server.start();
    }

    loadProto(path) {
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

    loadAllProtos(protoRoot) {
        const shopPackage = this.packages.shop.shop;

        this.server.addService(
            shopPackage.UserService.service,
            {}
        );

    }
}

// ShopServer.PORT = process.env.PORT;

module.exports = ShopServer;
