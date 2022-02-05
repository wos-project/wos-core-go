#! /bin/bash
S1=v1wosstage
S2=$HOME/${S1}
S3=/var/tmp/${S1}
S4=$HOME/backups/$S1/`date +"%Y%m%dT%H%M%S"`/
M1=wos-core-go

mkdir -p $S3
mkdir -p $S4
cp $S2/$M1 $S4

pkill -f ${S2}/config.yaml
cp $M1 $S2
nohup ${S2}/$M1 -log_dir=${S3} -config ${S2}/config.yaml >> ${S3}/${M1}.log &
