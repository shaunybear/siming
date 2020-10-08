cc_library(
    name = "mac",
    srcs = glob(["*.c", "region/*.c", "lmhandler/*.c", "lmhandler/packages/*.c", "soft-se/*.c"]),
    hdrs = glob(["*.h", "region/*.h", "lmhandler/*.h", "lmhandler/packages/*.h", "soft-se/*.h"]),
    copts = ["-Imac/region -Imac -Isystem -Iradio -Imac/lmhandler -Imac/lmhandler/packages \
              -Imac/soft-se -DSECURE_ELEMENT_PRE_PROVISIONED -DACTIVE_REGION=LORAMAC_REGION_US915"],
    deps = [ "//system:system", "//radio:radio"],
    visibility = ["//main:__pkg__"]
)