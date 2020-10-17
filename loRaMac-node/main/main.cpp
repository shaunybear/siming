#include <iostream>
#include <iterator>
#include <stdio.h>
#include <boost/program_options.hpp>

#include <czmq.h>
// #include <zmq.h>

const char* ENV_MAC_SERVICE_RPC_ADDR = "MAC_RPC_BACKEND_ADDRESS";

namespace po = boost::program_options;

using namespace std;

#include <boost/log/core.hpp>
#include <boost/log/trivial.hpp>
#include <boost/log/expressions.hpp>
#include <boost/log/sinks/text_file_backend.hpp>
#include <boost/log/utility/setup/file.hpp>
#include <boost/log/utility/setup/common_attributes.hpp>
#include <boost/log/sources/severity_logger.hpp>
#include <boost/log/sources/record_ostream.hpp>
#include <boost/log/utility/setup/console.hpp>

namespace logging = boost::log;
namespace src = boost::log::sources;
namespace sinks = boost::log::sinks;
namespace keywords = boost::log::keywords;

using namespace logging::trivial;


void init_logging()
{
    logging::register_simple_formatter_factory<logging::trivial::severity_level, char>("Severity");

    logging::add_file_log
    (
        keywords::file_name = "/opt/siming/var/log/sample_%N.log",                    /*< file name pattern >*/
        keywords::rotation_size = 10 * 1024 * 1024,                                   /*< rotate files every 10 MiB... >*/
        keywords::time_based_rotation = sinks::file::rotation_at_time_point(0, 0, 0), /*< ...or at midnight >*/
        keywords::format = "[%TimeStamp%] [%Severity%] %Message%",                    /*< log record format >*/
        keywords::auto_flush = true
    );

    logging::core::get()->set_filter
    (
        logging::trivial::severity >= logging::trivial::info
    );

    logging::add_console_log(std::cout, 
                             boost::log::keywords::format = ">> [%TimeStamp%] [%Severity%] %Message%",
                             boost::log::keywords::auto_flush = true);


    logging::add_common_attributes();
}


static void worker_task (const char* endpoint, const char *deveui, zsock_t *pipe, void *args)
{
    // Signal ready
    zsock_signal(pipe, 0);

    zsock_t *worker = zsock_new_req (endpoint);
    zpoller_t *poller = zpoller_new (pipe, worker, NULL);
    zpoller_set_nonstop(poller, true);

#define WORKER_READY "\001"
    //  Tell broker we're ready for work
    zframe_t *frame = zframe_new (WORKER_READY, 1);
    zframe_send (&frame, worker, 0);

    //  Process messages as they arrive
    while (true) {
        zsock_t *ready = zpoller_wait (poller, -1);
        if (ready == NULL) continue;   // Interrupted
        else if (ready == pipe) break; // Shutdown
        else assert(ready == worker);  // Data Available

        zmsg_t *msg = zmsg_recv (worker);
        if (!msg)
            break;              //  Interrupted
        zframe_print (zmsg_last (msg), "Worker: ");
        zframe_reset (zmsg_last (msg), "OK", 2);
        zmsg_send (&msg, worker);
    }

    zpoller_destroy(&poller);
    zsock_destroy(&worker);
}


static void start_mac_service(const char* endpoint, const char*deveui) {
    src::severity_logger< severity_level > lg;

    BOOST_LOG_SEV(lg, info) << "start_mac_services backend=" << ipc_address << " , identity=" << deveui;
    // zmq_setsockopt(socket, ZMQ_IDENTITY, deveui, strlen(deveui));
}




int main(int ac, char* av[])
{
    const char* endpoint = NULL;
    po::variables_map vm;        

    init_logging();

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

    if((endpoint = std::getenv(ENV_MAC_SERVICE_RPC_ADDR)) != NULL){
        std::cout <<  ENV_MAC_SERVICE_RPC_ADDR << " set to: " << endpoint << '\n';
    }
    else {
        cerr << ENV_MAC_SERVICE_RPC_ADDR << "is not set\n";
        return 1;
    } 

    if (vm.count("deveui")) {

        // start_mac_service(endpoint, vm["deveui"].as<string>().c_str());
        worker_task(endpoint, vm["deveui"].as<string>().c_str());
    }
    else{
        cerr << "Device EUI is not set\n";
        return 1;
    }

    return 0;
}