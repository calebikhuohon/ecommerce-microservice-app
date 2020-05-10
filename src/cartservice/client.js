const path = require('path');
const grpc = require('grpc');
const leftpad = require('left-pad');
const pino = require('pino');

const PROTO_PATH = path.join(__dirname, './proto/app.proto');
const PORT = process.env.PORT || "5150";

const shopProto = grpc.load(PROTO_PATH).shop;
const client = new shopProto.CartService(`localhost:${PORT}`,
    grpc.credentials.createInsecure());

const logger = pino({
    name: 'cartservice-client',
    messageKey: 'message',
    changeLevelName:'severity',
    useLevelLabels: true
});

client.AddItem({
    user_id: '1',
    item: {
        product_id: 'www',
        quantity: 2
    }
}, (err, response) => {
   if (err) {
       logger.error(`Error in AddItem 0: ${err}`);
   } else {
       console.log('add cart 0', response);
       logger.info(`data: ${response}`);
   }
});

client.AddItem({
    user_id: '2',
    item: {
        product_id: 'www',
        quantity: 2
    }
}, (err, response) => {
    if (err) {
        logger.error(`Error in AddItem 1: ${err}`);
    } else {
        console.log('add cart 1', response);
        logger.info(`data: ${response}`);
    }
});

client.EmptyCart({user_id: '1'}, (err, response) => {
   if (err) {
       logger.error(`Error in EmptyCart: ${err}`);
   }  else {
       console.log('empty cart', response);
       logger.info(`data: ${response}`);
   }
});

client.AddItem({
    user_id: '3',
    item: {
        product_id: 'www',
        quantity: 2
    }
}, (err, response) => {
    if (err) {
        logger.error(`Error in AddItem 2: ${err}`);
    } else {
        console.log('add item 2', response);
        logger.info(`data: ${response}`);
    }
});

const getReq = {
    user_id: '3'
}
client.GetCart(getReq, (err, response) => {
    if (err) {
        logger.error(`Error in GetCart: ${err}`);
    } else {
        console.log('get cart: ',response);
        logger.info(`data: ${response}`)
    }
})


