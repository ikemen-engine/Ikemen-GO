DEST=$(dirname -- $0)
unzip -o Ikemen_Mugen_Files.zip  -d ${DEST}
cp -r "${DEST}/Ikemen Mugen Files/" .
rm -rf "${DEST}/Ikemen Mugen Files"

echo "Installed at ${DEST}" >> ikemen_install.log

