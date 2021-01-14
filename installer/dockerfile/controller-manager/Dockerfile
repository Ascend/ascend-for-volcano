# Copyright 2019 The Volcano Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


FROM alpine:latest

COPY vc-controller-manager /vc-controller-manager

RUN adduser -h /home/hwMindX -u 9000 -D -s /bin/ash hwMindX && \
    chown -R hwMindX:hwMindX /vc-controller-manager && \
    chmod 500 /vc-controller-manager

USER hwMindX