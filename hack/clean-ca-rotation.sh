set -aux

rm -f ${CERTDIR}/ca.crt;
rm -f ${CERTDIR}/ca.key;
rm -f ${CERTDIR}/ca.crt.new;
rm -f ${CERTDIR}/ca.key.new;
mv -f ${CERTDIR}/ca.crt.old ${CERTDIR}/ca.crt;
mv -f ${CERTDIR}/ca.key.old ${CERTDIR}/ca.key;