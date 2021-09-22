## goscgi

SimpleCGI protocol implementation for Go lang. Allows creation of a basic HTTP server if used with Nginx or other SCGI capable web server.

### Nginx configuration

Locate Nginx configuration file. In Ubuntu it may be located at `/etc/nginx/sites-enabled/default`.
Add scgi_pass & include scgi_params directives in the root location.
~~~
location / {
	scgi_pass 127.0.0.1:8080;
	#scgi_pass unix:/tmp/goscgi.socket;
	include scgi_params;
}
~~~
If you use unix sockets, don't forget to give write permission
to www-data (default nginx user) on the socket file (created at runtime).
The examples below, use tcp sockets and don't need any special treatment.
Save the config file & restart the Nginx service. In Ubuntu: `sudo service nginx restart`.

### Usage

See `goscgi/benchmarks/test/main.go`.

### Benchmarking

Trying to see the performance of the SimpleCGI vs FastCGI vs Go Http server + Nginx proxy vs direct Go Http server,
 I performed some tests on my laptop Core2Duo P8400 2.6Ghz, using go version devel +3346bb37412c Fri Apr 05 13:43:18 2013 +1100 linux/amd64 and Apache Bench.
The code and screenshots are available in `/benchmarks`.
Nginx configuration used:
~~~
worker_processes: 4 (default)
worker_connections: 768 (default)

    location /scgi {
		scgi_pass 127.0.0.1:8080;
        include scgi_params;
    }
    location /fcgi {
        fastcgi_pass 127.0.0.1:8081;
        include fastcgi_params;
    }
    location /gosrv {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
    }
~~~ 
Apache Bench: `ab -n 2000 -c 50 http://localhost` + /scgi ,/fcgi, /gosrv respectively :8082/gosrv when we access Go HTTP Server directly.

*    ![SCGI vs FCGI vs proxied Go http server 1](https://github.com/leeview/goscgi/blob/master/benchmarks/ApacheBench_scgi_fcgi_gosrv_with_proxy1.png)
*    ![SCGI vs FCGI vs proxied Go http server 2](https://github.com/leeview/goscgi/blob/master/benchmarks/ApacheBench_scgi_fcgi_gosrv_with_proxy2.png)
*    ![SCGI vs FCGI vs proxied Go http server 3](https://github.com/leeview/goscgi/blob/master/benchmarks/ApacheBench_scgi_fcgi_gosrv_with_proxy3.png)
*    ![SCGI vs FCGI vs direct Go http server 1](https://github.com/leeview/goscgi/blob/master/benchmarks/ApacheBench_scgi_fcgi_gosrv_with_no_proxy1.png)
*    ![SCGI vs FCGI vs direct Go http server 2](https://github.com/leeview/goscgi/blob/master/benchmarks/ApacheBench_scgi_fcgi_gosrv_with_no_proxy2.png)


