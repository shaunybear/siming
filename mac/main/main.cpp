#include <iostream>
#include <iterator>
#include <stdio.h>
#include <boost/program_options.hpp>

#include <zmq.h>

namespace po = boost::program_options;
namespace asio = boost::asio;

using namespace std;

int hello_server(const char* server_address) {
    //  Socket to talk to clients
    void *context = zmq_ctx_new ();
    void *responder = zmq_socket (context, ZMQ_REP);
    int rc = zmq_bind (responder, "tcp://*:5555");
    assert (rc == 0);

    while (1) {
        char buffer [10];
        zmq_recv (responder, buffer, 10, 0);
        printf ("Received Hello\n");
        sleep (1);          //  Do some 'work'
        zmq_send (responder, "World", 5, 0);
    }
    return 0;
}

int hello_client(const char* server_address) {
    printf ("Connecting to hello world server…\n");
    void *context = zmq_ctx_new ();
    void *requester = zmq_socket (context, ZMQ_REQ);
    zmq_connect (requester,  "tcp://localhost:5555");

    int request_nbr;
    for (request_nbr = 0; request_nbr != 10; request_nbr++) {
        char buffer [10];
        printf ("Sending Hello %d…\n", request_nbr);
        zmq_send (requester, "Hello", 5, 0);
        zmq_recv (requester, buffer, 10, 0);
        printf ("Received World %d\n", request_nbr);
    }
    zmq_close (requester);
    zmq_ctx_destroy (context);
    return 0;
}


int main(int ac, char* av[])
{
    bool server_mode = false;
    const char* server_address = NULL;

    try {
        po::options_description desc("Options");
        desc.add_options()
            ("help, h", "Help screen")
            ("config", po::value<string>(), "Config file")
            ("server,s", "Server mode");

        po::variables_map vm;        
        po::store(po::parse_command_line(ac, av, desc), vm);
        po::notify(vm);    

        if (vm.count("help")) {
            cout << desc << "\n";
            return 0;
        }

        if (vm.count("config")) {
            cout << "Config file set to " 
                 << vm["config"].as<std::string>() << ".\n";
        } else {
            cout << "Config file not set.\n";
            // return 1;
        }

        if(vm.count("server")) {
            server_mode =true;
        }
    }
    catch(exception& e) {
        cerr << "error: " << e.what() << "\n";
        return 1;
    }
    catch(...) {
        cerr << "Exception of unknown type!\n";
    }

    if((server_address = std::getenv("ZMQ_DEVICE_SERVER")) != NULL){
        std::cout << "ZMQ Device Server is: " << server_address << '\n';
    }
    else {
        cerr << "ZMQ_DEVICE_SERVER is not set!\n";
        return 1;
    }

    if(server_mode){
        hello_server(server_address);
    }
    else{
        hello_client(server_address);
    }

    return 0;
}