#!/bin/bash

REMOTE=""


curl="http "

function change_setting() {
  $curl PUT $REMOTE/api/v2/img/$1?vValue=$2
}

change_setting("ASTV_logo", )
# "ASTV_logo": "/uploads/grid/uploads/upload/file/565caac7de467c0011000000/astv_tile.svg",
# "DIA_logo": "/assets/gallery/dia_tile.svg",
# "DOS_logo": "/assets/gallery/dos_tile.svg",
# "DWB_Logo": "/uploads/grid/uploads/upload/file/565c43fd9ec12f0011000000/DoctorsWithoutBorders_final.png",
# "ICRC_Logo": "/uploads/grid/uploads/upload/file/565c4f69d805600011000003/ICRC_final.png",
# "MC_Logo": "/uploads/grid/uploads/upload/file/565c44fa9ec12f0011000006/MercyCorps_final.png",
# "NGIA_logo": "/uploads/grid/uploads/upload/file/565c4cad89e5140011000000/NGIA_final.png",
# "RDP_logo": "/uploads/grid/uploads/upload/file/565c4f54d805600011000000/RPD_final.png",
# "SAPA_logo": "/uploads/grid/uploads/upload/file/565c4dc889e5140011000006/sapa_final.png",
# "TPP_logo": "/uploads/grid/uploads/upload/file/565c4d2289e5140011000003/TPP_final.png",
# "UN_Logo": "/uploads/grid/uploads/upload/file/565c45719ec12f0011000009/NATO_final.png",
# "USAID_logo": "/uploads/grid/uploads/upload/file/565c4e0389e5140011000009/USAID_final.png",
# "VATCLogo_black": "/uploads/grid/uploads/upload/file/566568e0ba077b001a000000/5a2b5fc97e.png",
# "VATClogo": "/uploads/grid/uploads/upload/file/565c48009ec12f0011000015/VATC_Logo_light.png",
# "WW_Logo": "/uploads/grid/uploads/upload/file/565c44c59ec12f0011000003/WorldVision_final.png",
# "favicon": "http://sitrep.vatcinc.com/favicon.png",
# "intellipedia_logo": "/uploads/grid/uploads/upload/file/565787e21ca46b0018000004/220px-Intellipedia_Logo.jpg",
# "logo": "/uploads/grid/uploads/upload/file/565c485d9ec12f0011000018/SITREP_logo.png",
# "logo_dark": "/uploads/grid/uploads/upload/file/565cabfc73392c0011000000/SITREP_logo.png"
