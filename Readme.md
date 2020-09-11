**GoLang Codig Challenge**

This is a simple HTTP REST endpoint which takes a URL encoded IP address against a known list of whitelisted countries 
and responds as to whether that IP is authorized to continue. The service uses GeoLite2 data to determine IP address origin.

If the geoipupdate utility is installed the default installation of this service will update the built-in version of 
the GeoLite2-Country.mmdb data set.

**Usage**

Submit a POST request to _http:\\\\[hostname]:10000\processIpRequest_ with the following format in the body: 

`{
   "request":"8.8.8.8",
   "whitelist": ["United States","Canada"]
 }`
 
This product includes GeoLite2 data created by MaxMind, available from
<a href="https://www.maxmind.com">https://www.maxmind.com</a>.
