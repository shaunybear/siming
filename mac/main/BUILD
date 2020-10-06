cc_binary(
    name = "main",
    srcs = ["main.cpp"],
    deps = ["//mac:mac"],
    copts =["-Imac -Imac/lmhandler/packages -Imac/lmhandler -Isystem -Iradio"],
    linkopts = ["-lzmq -lboost_system -lboost_regex -lboost_program_options -lpthread"]
)