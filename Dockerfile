FROM golang as build-env

# TODO: Can be optimized
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go get ./... && go build .

CMD ["bash"]

# When using Kubernetes:
#
# kubectl exec raftbench0 -c bench -- ./raftbench --engine etcd --mil 1000 --firstWait 20000 --step 1 --id 1 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 --test --logfile etcd.csv"]
# kubectl exec raftbench1 -c bench -- ./raftbench --engine etcd --id 2 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 " &
# kubectl exec raftbench3 -c bench -- ./raftbench --engine etcd --id 3 --cluster http://raft0:12379,http://raft1:12379,http://raft2:12379 "

# Then inspect the log csv-file:
# kubectl exec -ti raftbench0 -c bench -- bash"
