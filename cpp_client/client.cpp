#include <iostream>
#include <string>
#include <vector>
#include <sstream>
#include <ctime>

#include <iomanip>
#include <openssl/hmac.h>
#include <boost/uuid/uuid.hpp>
#include <boost/uuid/uuid_generators.hpp>
#include <boost/uuid/uuid_io.hpp>
#include <sio_client.h>

#include <openssl/bio.h>
#include <openssl/evp.h>
#include <openssl/buffer.h>
#include <chrono>

//prod keys
const std::string API_KEY = "*";
const std::string PASSPHRASE = "*";
const std::string SECRET_KEY = "*";
const std::string URL = "https://stream.falconx.io";


std::string base64_encode(const unsigned char* input, size_t len)
{
    BIO *bmem = NULL, *b64 = NULL;
    BUF_MEM *bptr = NULL;

    b64 = BIO_new(BIO_f_base64());
    bmem = BIO_new(BIO_s_mem());
    b64 = BIO_push(b64, bmem);
    BIO_write(b64, input, len);
    BIO_flush(b64);
    BIO_get_mem_ptr(b64, &bptr);

    std::string result(bptr->data, bptr->length-1);
    BIO_free_all(b64);

    return result;
}

std::vector<std::uint8_t> base64_decode(const std::string& input) {
    BIO *bio_mem, *bio_b64;
    BUF_MEM *buf_mem_ptr;

    bio_mem = BIO_new_mem_buf(input.data(), -1);
    bio_b64 = BIO_new(BIO_f_base64());
    BIO_set_flags(bio_b64, BIO_FLAGS_BASE64_NO_NL);

    // Chain bio_mem and bio_b64
    bio_mem = BIO_push(bio_b64, bio_mem);

    std::vector<std::uint8_t> output(input.size());
    int decoded_size = BIO_read(bio_mem, output.data(), input.size());

    // Cleanup
    BIO_free_all(bio_mem);

    if (decoded_size < 0) {
        throw std::runtime_error("Error: Failed to decode base64 string.");
    }

    output.resize(decoded_size);
    return output;
}

std::string get_current_time_in_ms() {
    using namespace std::chrono;
    auto now = system_clock::now();
    auto duration_since_epoch = now.time_since_epoch();
    auto micros = duration_cast<microseconds>(duration_since_epoch).count();
    std::ostringstream oss;
    oss << std::fixed << std::setprecision(6) << static_cast<double>(micros) / 1000000;
    return oss.str();
}

std::map<std::string, std::string> create_header(const std::string &api_key, const std::string &secret_key, const std::string &passphrase) {
    std::string timestamp = get_current_time_in_ms();
    std::string message = timestamp + "GET/socket.io/";

    std::vector<std::uint8_t> decoded_data = base64_decode(secret_key);
    std::string hmac_key(decoded_data.begin(), decoded_data.end());

    unsigned char* signature = HMAC(EVP_sha256(), hmac_key.c_str(), hmac_key.size(),
                                (unsigned char *)message.data(), message.size(),
                                nullptr, nullptr);

    std::string signature_b64 = base64_encode(signature, 32);

    return {
        {"FX-ACCESS-SIGN", signature_b64},
        {"FX-ACCESS-TIMESTAMP", timestamp},
        {"FX-ACCESS-KEY", api_key},
        {"FX-ACCESS-PASSPHRASE", passphrase},
        {"Content-Type", "application/json"}
    };
}

void printRes(sio::message::ptr msg) {
    switch (msg->get_flag())
    {
        case sio::message::flag_string:
            std::cout << msg->get_string();
            break;

        case sio::message::flag_object:
            // Handle object messages
            // You can use msg->get_map() to access the key-value pairs
            for (const auto& element : msg->get_map()) {
                std::cout << element.first << ": ";
                printRes(element.second);
            }
            break;

        case sio::message::flag_array:
            // Handle array messages
            // You can use msg->get_vector() to access the elements
            for (const auto& element : msg->get_vector()) {
                printRes(element);
                std::cout<<",";
            }
            break;

        case sio::message::flag_integer:
            std::cout << msg->get_int();
            break;

        case sio::message::flag_double:
            std::cout << msg->get_double();
            break;

        case sio::message::flag_null:
            std::cout << "null";
            break;

        case sio::message::flag_boolean:
            std::cout << msg->get_bool();
            break;

        default:
            std::cout << "unsupported type";
            break;
    }
    std::cout << "; ";
}

class FastRFSClient
{
public:
    FastRFSClient(sio::client *socket_io_client, const std::string &namespace_)
        : socket_io_client(socket_io_client), namespace_(namespace_) {}

    void OnConnected()
    {
        std::cout << "Success: Connected to the server" << std::endl;
        for (const auto &subscription_request : subscription_requests)
        {
            socket_io_client->socket(namespace_)->emit("subscribe", subscription_request);
        }
        std::cout << "Finished subscribing." << std::endl;
    }

    void populate_subscription_requests(const std::vector<std::string> &token_pairs, const std::vector<int> &levels)
    {
        for (const auto &token_pair : token_pairs)
        {
            std::istringstream iss(token_pair);
            std::string base_token, quote_token;
            getline(iss, base_token, '/');
            getline(iss, quote_token, '/');

            sio::message::ptr subscription_request = sio::object_message::create();
            sio::message::ptr token_pair_object = sio::object_message::create();
            token_pair_object->get_map()["base_token"] = sio::string_message::create(base_token);
            token_pair_object->get_map()["quote_token"] = sio::string_message::create(quote_token);
            subscription_request->get_map()["token_pair"] = token_pair_object;
            sio::message::ptr quantity_array = sio::array_message::create();
            for (int level : levels) {
                quantity_array->get_vector().push_back(sio::int_message::create(level));
            }
            subscription_request->get_map()["quantity"] = quantity_array;
            subscription_request->get_map()["quantity_token"] = sio::string_message::create(base_token);
            subscription_request->get_map()["client_request_id"] = sio::string_message::create(boost::uuids::to_string(boost::uuids::random_generator()()));
            subscription_request->get_map()["echo_id"] = sio::bool_message::create(true);

            subscription_requests.push_back(subscription_request);
        }
    }

    void connect()
    {
        std::map<std::string, std::string> headers = create_header(API_KEY, SECRET_KEY, PASSPHRASE);

        socket_io_client->connect(URL, {}, headers);

        std::cout<<"Namespace is: "<<namespace_<<"\n";

        socket_io_client->socket(namespace_)->on("connect", [&](sio::event &)
                            {
                                std::cout << "Server connected." << std::endl;
                                for (const auto &subscription_request : subscription_requests)
                                {
                                    socket_io_client->socket(namespace_)->emit("subscribe", subscription_request);
                                }
                                std::cout << "Finished subscribing." << std::endl;
                            });

        socket_io_client->socket(namespace_)->on("disconnect", [&](sio::event &e)
                                                 {
                                                    std::cout << "Connect Error: ";
                                                    sio::message::ptr msg = e.get_message();
                                                    printRes(msg);
                                                    std::cout << std::endl << std::endl;
                                                 });

        socket_io_client->socket(namespace_)->on("connect_error", [&](sio::event &e)
                                                 {
                                                    std::cout << "Connect Error: ";
                                                    sio::message::ptr msg = e.get_message();
                                                    printRes(msg);
                                                    std::cout << std::endl << std::endl;
                                                 });

        socket_io_client->socket(namespace_)->on("response", [&](sio::event &e)
        {
            std::cout << "Response args: ";

            sio::message::ptr msg = e.get_message();
            printRes(msg);

            std::cout << std::endl << std::endl;
        });


        socket_io_client->socket(namespace_)->on("stream", [&](sio::event &e)
                                                 {
                                                     std::cout << "Streaming args: ";
                                                     sio::message::ptr msg = e.get_message();
                                                     printRes(msg);
                                                     std::cout << std::endl << std::endl;
                                                 });

        socket_io_client->socket(namespace_)->on("error", [&](sio::event &e)
                                                 {
                                                    std::cout << "Error: ";
                                                    sio::message::ptr msg = e.get_message();
                                                    printRes(msg);
                                                    std::cout << std::endl << std::endl;
                                                 });

        // The following line blocks until the connection is closed.

        socket_io_client->sync_close();
    }

private:
    sio::client *socket_io_client;
    std::string namespace_;
    std::vector<sio::message::ptr> subscription_requests;
};

void make_new_connection(const std::vector<std::string> &token_pairs, const std::vector<int> &levels)  {
    sio::client socket_io_client;

    FastRFSClient client(&socket_io_client, "/streaming");
    client.populate_subscription_requests(token_pairs, levels);
    // Set up event handlers
    socket_io_client.set_open_listener(std::bind(&FastRFSClient::OnConnected, client));
    socket_io_client.set_reconnect_attempts(0); // set this to zero so it doesnt reconnect with old headers automatically
    client.connect();
    // the code flow comes here when connection is closed.
}


int main(int argc, char **argv)
{
    std::vector<std::string> token_pairs = {"BTC/USD"};
    std::vector<int> levels = {30};

    if (argc > 1)
    {
        std::istringstream iss_token_pairs(argv[1]);
        std::string token_pair;
        token_pairs.clear();
        while (getline(iss_token_pairs, token_pair, ','))
        {
            token_pairs.push_back(token_pair);
        }
    }

    if (argc > 2)
    {
        std::istringstream iss_levels(argv[2]);
        std::string level_str;
        levels.clear();
        while (getline(iss_levels, level_str, ','))
        {
            int level = std::stoi(level_str);
            levels.push_back(level);
        }
    }

    while(true){
        make_new_connection(token_pairs, levels);
        std::cout << "Waiting for 5 seconds before reconnecting "<<std::endl;
        std::this_thread::sleep_for(std::chrono::seconds(5));
        std::cout << "ReConeccting: "<<std::endl;
    }

    return 0;
}
