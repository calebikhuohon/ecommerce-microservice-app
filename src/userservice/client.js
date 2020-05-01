const  path = require('path');
const grpc = require('grpc');
const leftpad = require('left-pad');
const pino = require('pino');

const PROTO_PATH = path.join(__dirname, './proto/app.proto');
const PORT = 7000;

const shopProto = grpc.load(PROTO_PATH).shop;
const client = new shopProto.UserService(`localhost:${PORT}`,
    grpc.credentials.createInsecure());

const logger = pino({
    name: 'userservice-client',
    messageKey: 'message',
    changeLevelName: 'severity',
    useLevelLabels: true
});

const request = {
    value: 'enter_the_user_id_to_be_requested',
};

client.getUser({}, (err, response) => {
    if (err) {
        logger.error(`Error in getUser: ${err}`);
    } else {
        logger.info(`user: ${response.firstName} ${response.lastName}`);
    }
})


