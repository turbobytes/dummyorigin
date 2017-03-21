#!/bin/bash

mkdir -p assets

#http://stackoverflow.com/a/14371026
declare -A arr

arr["assets/15kb.png"]="https://upload.wikimedia.org/wikipedia/en/6/66/Circle_sampling.png"
arr["assets/15kb.jpg"]="http://static.cdnplanet.com/static/rum/15kb-image.jpg"
arr["assets/100kb.jpg"]="http://static.cdnplanet.com/static/rum/100kb-image.jpg"
arr["assets/10kb.js"]="https://rum.turbobytes.com/static/rum/rum.js"
arr["assets/160kb.js"]="https://ajax.googleapis.com/ajax/libs/angularjs/1.5.7/angular.min.js"
arr["assets/86kb.js"]="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"
arr["assets/100kb.js"]="https://cdn.jsdelivr.net/angular.bootstrap/2.5.0/ui-bootstrap.min.js"
arr["assets/10mb.mp4"]="https://tdispatch.com/wp-content/uploads/2014/11/tdispatch-10MB-MP4-.mp4?_=2"
#arr["assets/30mb.mp4"]="http://www.sample-videos.com/video/mp4/480/big_buck_bunny_480p_30mb.mp4"
arr["assets/150mb.avi"]="http://download.blender.org/peach/bigbuckbunny_movies/big_buck_bunny_480p_stereo.avi"

for key in ${!arr[@]}; do
    echo ${key} ${arr[${key}]}
	if [ -f ${key} ]; then
		echo "File already exists"
	else
		curl -o ${key} ${arr[${key}]}
	fi
done
