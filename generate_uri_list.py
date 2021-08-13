import os
import json

files = os.listdir("./Templates")

function_blob = {}

for file in files:
    with open(f"./Templates/{file}") as template_file:
        data = json.load(template_file)
        
    function_name = file.split("-")[0]

    try:
        uri = data["Resources"][function_name]["Properties"]['CodeUri']
        function_blob[function_name] = uri
    except Exception as e:
        pass

with open('function_uris.json', 'w', encoding='utf-8') as function_uris:
    json.dump(function_blob, function_uris, ensure_ascii=False, indent=4)

