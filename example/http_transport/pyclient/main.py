import test
import sys
import pprint
from urlparse import urlparse
from thrift.transport import TTransport
from thrift.transport import THttpClient
from thrift.protocol import TBinaryProtocol

from test import test
from test.ttypes import *

pp = pprint.PrettyPrinter(indent = 2)
# for goserver
# transport = THttpClient.THttpClient("http://127.0.0.1:19090/")
# for goserver and goserver2
transport = THttpClient.THttpClient("http://127.0.0.1:19090/api")
protocol = TBinaryProtocol.TBinaryProtocol(transport)
client = test.Client(protocol)
transport.open()
pp.pprint(client.hello("world",))
transport.close()

