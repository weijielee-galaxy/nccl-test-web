 /usr/local/sihpc/bin/mpirun \
     --allow-run-as-root \
     --hostfile ./iplist \
     --map-by ppr:8:node  \
     --mca oob_tcp_if_include bond0 \
     --mca pml ^ucx   \
     --mca btl self,tcp \
     --mca btl_tcp_if_include bond0   \
     --mca routed direct \
     --mca plm_rsh_no_tree_spawn 1 \
     -x UCX_TLS=tcp \
     -x NCCL_DBEUG=INFO \
     -x NCCL_IB_GID_INDEX=3 \
     -x NCCL_MIN_NCHANNELS=32 \
     -x NCCL_IB_QPS_PER_CONNECTION=8 \
     /usr/local/sihpc/libexec/nccl-tests/nccl_test  -b 1 -e 1k

