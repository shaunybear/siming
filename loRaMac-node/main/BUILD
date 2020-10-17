cc_binary(
    name = "loRaMac-node",
    srcs = ["main.cpp"],
    deps = ["//mac:mac"],
    copts =["-Imac -Imac/lmhandler/packages -Imac/lmhandler -Isystem -Iradio -DBOOST_LOG_DYN_LINK"],
    linkopts = ["-lzmq -lboost_system -lboost_log -lboost_thread -lboost_regex -lboost_program_options -lpthread -lboost_log_setup"]
)