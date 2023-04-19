set -aux

mv ${CERTDIR}/ca.crt ${CERTDIR}/ca.crt.old
mv ${CERTDIR}/ca.key ${CERTDIR}/ca.key.old
kubeadm init --cert-dir=${CERTDIR}  phase certs ca || echo "kubeadm init didn't exit with 0"
cp ${CERTDIR}/ca.crt ${CERTDIR}/ca.crt.new
cp ${CERTDIR}/ca.key ${CERTDIR}/ca.key.new
cat ${CERTDIR}/ca.crt.new ${CERTDIR}/ca.crt.old > ${CERTDIR}/ca.crt
cp ${CERTDIR}/ca.crt ${CERTDIR}/ca.crt.bundle
base64_encoded_ca=$(base64 -w0 ${CERTDIR}/ca.crt)
for namespace in $( kubectl get namespace --no-headers -o name | cut -d / -f 2 ); do
    for token in $( kubectl get secrets --namespace "$namespace" --field-selector type=kubernetes.io/service-account-token -o name); do
    kubectl get $token --namespace "$namespace" -o yaml | \
        /bin/sed "s/\(ca.crt:\).*/\1 $base64_encoded_ca/" | \
    kubectl apply -f - ;
    done;
done