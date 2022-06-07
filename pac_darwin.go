package ieproxy

/*
#cgo LDFLAGS: -framework CoreFoundation
#cgo LDFLAGS: -framework CFNetwork
#include <strings.h>
#include <CFNetwork/CFProxySupport.h>

#define STR_LEN 128

void proxyAutoConfCallback(void* client, CFArrayRef proxies, CFErrorRef error) {
	CFTypeRef* result_ptr = (CFTypeRef*)client;
	if (error != NULL) {
		*result_ptr = CFRetain(error);
	  } else {
		*result_ptr = CFRetain(proxies);
	  }
	  CFRunLoopStop(CFRunLoopGetCurrent());
}

int intCFNumber(CFNumberRef num) {
	int ret;
	CFNumberGetValue(num, kCFNumberIntType, &ret);
	return ret;
}

char* _getProxyUrlFromPac(char* pac, char* reqCs) {
	printf("CGO:_getProxyUrlFromPac:pac:%s\n", pac);
	printf("CGO:_getProxyUrlFromPac:reqCs:%s\n", reqCs);

	char* retCString = (char*)calloc(STR_LEN, sizeof(char));

	CFStringRef reqStr = CFStringCreateWithCString(NULL, reqCs, kCFStringEncodingUTF8);
	printf("CGO:_getProxyUrlFromPac:reqStr:%p\n", reqCs);
	CFStringRef pacStr = CFStringCreateWithCString(NULL, pac, kCFStringEncodingUTF8);
	printf("CGO:_getProxyUrlFromPac:pacStr:%p\n", pacStr);
	CFURLRef pacUrl = CFURLCreateWithString(NULL, pacStr, NULL);
	printf("CGO:_getProxyUrlFromPac:pacUrl:%p\n", pacUrl);
	CFURLRef reqUrl = CFURLCreateWithString(NULL, reqStr, NULL);
	printf("CGO:_getProxyUrlFromPac:reqUrl:%p\n", reqUrl);

	CFTypeRef result = NULL;
	CFStreamClientContext context = { 0, &result, NULL, NULL, NULL };
	CFRunLoopSourceRef runloop_src = CFNetworkExecuteProxyAutoConfigurationURL(pacUrl, reqUrl, proxyAutoConfCallback, &context);
	printf("CGO:_getProxyUrlFromPac:runloop_src:%p\n", runloop_src);

	if (runloop_src) {
		const CFStringRef private_runloop_mode = CFSTR("go-ieproxy");
		CFRunLoopAddSource(CFRunLoopGetCurrent(), runloop_src, private_runloop_mode);
		CFRunLoopRunInMode(private_runloop_mode, DBL_MAX, false);
		CFRunLoopRemoveSource(CFRunLoopGetCurrent(), runloop_src, kCFRunLoopCommonModes);

		if (CFGetTypeID(result) == CFArrayGetTypeID()) {
			CFArrayRef resultArray = (CFTypeRef)result;
			if (CFArrayGetCount(resultArray) > 0) {
				CFDictionaryRef pxy = (CFDictionaryRef)CFArrayGetValueAtIndex(resultArray, 0);
				CFStringRef pxyType = CFDictionaryGetValue(pxy, kCFProxyTypeKey);

				if (CFEqual(pxyType, kCFProxyTypeNone)) {
					// noop
				}

				if (CFEqual(pxyType, kCFProxyTypeHTTP)) {
					CFStringRef host = (CFStringRef)CFDictionaryGetValue(pxy, kCFProxyHostNameKey);
					CFNumberRef port = (CFNumberRef)CFDictionaryGetValue(pxy, kCFProxyPortNumberKey);

					char host_str[STR_LEN - 16];
					CFStringGetCString(host, host_str, STR_LEN - 16, kCFStringEncodingUTF8);

					int port_int = 80;
					if (port) {
						CFNumberGetValue(port, kCFNumberIntType, &port_int);
					}

					sprintf(retCString, "%s:%d", host_str, port_int);
				}
			}
		} else {
			// error
		}
	}

	CFRelease(result);
	CFRelease(reqStr);
	CFRelease(reqUrl);
	CFRelease(pacStr);
	CFRelease(pacUrl);
	return retCString;
}

char* _getPacUrl() {
	char* retCString = (char*)calloc(STR_LEN, sizeof(char));
	CFDictionaryRef proxyDict = CFNetworkCopySystemProxySettings();
	CFNumberRef pacEnable = (CFNumberRef)CFDictionaryGetValue(proxyDict, kCFNetworkProxiesProxyAutoConfigEnable);

	if (pacEnable && intCFNumber(pacEnable)) {
		CFStringRef pacUrlStr = (CFStringRef)CFDictionaryGetValue(proxyDict, kCFNetworkProxiesProxyAutoConfigURLString);
		if (pacUrlStr) {
			CFStringGetCString(pacUrlStr, retCString, STR_LEN, kCFStringEncodingUTF8);
		}
	}

	CFRelease(proxyDict);
	return retCString;
}

*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

func (psc *ProxyScriptConf) findProxyForURL(URL string) string {
	if !psc.Active {
		return ""
	}
	proxy := getProxyForURL(psc.PreConfiguredURL, URL)
	return proxy
}

func getProxyForURL(pacFileURL, url string) string {
	if pacFileURL == "" {
		pacFileURL = getPacUrl()
	}
	if strings.TrimSpace(pacFileURL) == "" {
		return ""
	}

	csUrl := C.CString(url)
	csPac := C.CString(pacFileURL)
	fmt.Println()
	fmt.Println("#getProxyForURL:pacFileURL:", pacFileURL)
	fmt.Println("#getProxyForURL:url:", url)
	fmt.Println("#getProxyForURL:csUrl:", csUrl)
	fmt.Println("#getProxyForURL:csPac:", csPac)
	csRet := C._getProxyUrlFromPac(csPac, csUrl)

	defer C.free(unsafe.Pointer(csUrl))
	defer C.free(unsafe.Pointer(csPac))
	defer C.free(unsafe.Pointer(csRet))

	return C.GoString(csRet)
}

func getPacUrl() string {
	csRet := C._getPacUrl()

	defer C.free(unsafe.Pointer(csRet))
	return C.GoString(csRet)
}
