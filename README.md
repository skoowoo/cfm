cfm
===

对模块化程序设计更加友好的一个配置文件库

配置demo:

	error_log 			"/tmp/error.log";
	error_log_level  	error;
	
	tcp {
		tcp_nodelay on;
		
		http {
			listen 80;
			add_header "name=value";
		}
	}
	
	udp {
		reuseport on;
		
		dns {
			listen 53;
		}
	}