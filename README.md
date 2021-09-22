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
