#
# Copyright 2016 Planet Labs
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

import netaddr


def subnet(network_cidr, subnet_size, subnet_offset=0):
    net = netaddr.IPNetwork(network_cidr)
    subnets = list(net.subnet(subnet_size))
    return subnets[subnet_offset]


class FilterModule(object):
    def filters(self):
        return {'subnet': subnet}
