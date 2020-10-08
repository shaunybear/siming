#include <iostream>
#include <iterator>
#include <stdio.h>
#include <boost/program_options.hpp>

#include <zmq.h>

const char* ENV_MAC_SERVICE_RPC_ADDR = "MAC_RPC_BACKEND_ADDRESS";

namespace po = boost::program_options;

using namespace std;

static int start_mac_service(const char* ipc_address, const char*deveui) {
    //  Socket to talk to clients
    void *context = zmq_ctx_new ();
    void *responder = zmq_socket (context, ZMQ_REQ);

    printf("start_mac_service size=%ld %s\n", strlen(deveui),deveui);
    zmq_setsockopt(responder, ZMQ_IDENTITY, deveui, strlen(deveui));
    int rc = zmq_connect (responder, ipc_address); 
    assert (rc == 0);

    // Send ready signal
    zmq_send (responder, "\001", 1, 0);

    while (1) {
        char buffer [10];
        zmq_recv (responder, buffer, 10, 0);
        printf ("Received Hello\n");
        sleep (1);          //  Do some 'work'
        zmq_send (responder, "World", 5, 0);
    }
    return 0;
}

int main(int ac, char* av[])
{
    const char* ipc_address = NULL;
    po::variables_map vm;        

    try {
        po::options_description desc("Options");
        desc.add_options()
            ("help, h", "Help screen")
            ("deveui", po::value<string>(), "Device EUI ");

        po::store(po::parse_command_line(ac, av, desc), vm);
        po::notify(vm);    

        if (vm.count("help")) {
            cout << desc << "\n";
            return 1;
        }
    }
    catch(exception& e) {
        cerr << "error: " << e.what() << "\n";
        return 1;
    }
    catch(...) {
        cerr << "Exception of unknown type!\n";
    } 

    if((ipc_address = std::getenv(ENV_MAC_SERVICE_RPC_ADDR)) != NULL){
        std::cout <<  ENV_MAC_SERVICE_RPC_ADDR << " set to: " << ipc_address << '\n';
    }
    else {
        cerr << ENV_MAC_SERVICE_RPC_ADDR << "is not set\n";
        return 1;
    } 

    if (vm.count("deveui")) {

        start_mac_service(ipc_address, vm["deveui"].as<string>().c_str());
    }
    else{
        cerr << "Device EUI is not set\n";
        return 1;
    }

    return 0;
}