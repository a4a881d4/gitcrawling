#coding:utf-8
import argparse
import sys,os
pt = os.getcwd()+'/utils/dbrestful'
print(pt)
sys.path.append(pt)
from dbapi import web_app
from dbapi.app import define_urls


def get_options():
    parser = argparse.ArgumentParser()
    parser.add_argument('-d', '--debug', action='store_true', default=False)
    parser.add_argument('-p', '--port', type=int, default=7532)
    parser.add_argument('-b', '--database')
    parser.add_argument('-H', '--host', default='127.0.0.1')
    return parser.parse_args()


def configure_app(app, options):
    app.config['DB'] = options.database
    define_urls(app)


def dev_server():
    options = get_options()
    configure_app(web_app, options)
    web_app.run(debug=options.debug, port=options.port, host=options.host)


def run_server():
    options = get_options()
    configure_app(web_app, options)

    from gevent.pywsgi import WSGIServer

    server = WSGIServer((options.host, options.port), web_app)
    server.serve_forever()


if __name__ == '__main__':

    run_server()
