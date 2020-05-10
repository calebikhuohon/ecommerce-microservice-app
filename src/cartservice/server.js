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

//In-memory array of a user's cart items
//  Cart = [{
//        user_id: userId,
//        items: [{
//          cartItems
//     }]
//  }]
let Cart = [];

/**
 *
 * @param call {AddItemRequest }
 * @param callback { err, {} }
 */
function AddItem(call, callback) {
    let request = call.request;


    try {
        console.log(request.user_id, request.item);

        const cart = Cart.filter(item => {
            return cart.user_id === request.user_id;
        });

        if (cart.length === 0) {
            const cartItem = {
                user_id: request.user_id,
                items: [request.item]
            };
            Cart.push(cartItem);
        }

        if (cart.length === 1) {
            cart.items.push(request.item);
        }

        console.log('Cart:---------',Cart);

        callback(null, {});
    } catch (e) {
        logger.error(`adding item failed: ${e}`);
        callback(e);
    }
}

/**
 *
 * @param call { EmptyCartRequest }
 * @param callback {err, {} }
 * @constructor
 */
function EmptyCart(call, callback) {
    const request = call.request;

    try {
        Cart = Cart.filter(item => {
            return item.user_id !== request.user_id;
        });

        callback(null, {});
    }catch (e) {
        logger.error(`emptying cart with user id ${request.user_id} failed: ${e}`)
        callback(e);
    }
}

/**
 *
 * @param call { GetCartRequest }
 * @param callback { err, Cart }
 * @constructor
 */
function GetCart(call, callback) {
    const request = call.request;

    try {
        const userCart = Cart.filter(item => {
            return item.user_id === request.user_id;
        });

        callback(null, userCart);
    } catch (e) {

    }
}

/**
 * starts an RPC server that receives requests for the
 * User service at the sample server port
 */
function main() {
    logger.info(`starting gRPC server on port ${PORT}...`);
    const server = new grpc.Server();
    server.addService(shopProto.CartService.service, {AddItem, EmptyCart, GetCart});
    server.bind(`0.0.0.0:${PORT}`, grpc.ServerCredentials.createInsecure());
    server.start();
}

main();
