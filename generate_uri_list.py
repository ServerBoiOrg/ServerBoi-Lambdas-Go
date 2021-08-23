import os
import json

files = os.listdir("./Templates")

function_blob = {}

print(f"Templates in directory: {files}")

for file in files:
    print(f"Working on {file}")
    with open(f"./Templates/{file}") as template_file:
        data = json.load(template_file)
        
    function_name = file.split("-")[0]

    try:
        uri = data["Resources"][function_name]["Properties"]['CodeUri']
        split_uri = uri.split("/")[2:5]
        bucket = split_uri[0]
        function_blob[function_name] = {
            "bucket": bucket,
            "key": f"{split_uri[1]}/{split_uri[2]}",
        }
        print(f"Added {function_name} to uri list")
    except Exception as e:
        print(f"Error adding {function_name} to uri list.")
        pass

with open('function_uris.json', 'w', encoding='utf-8') as function_uris:
    json.dump(function_blob, function_uris, ensure_ascii=False, indent=4)

