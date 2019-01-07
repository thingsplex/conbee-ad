import os, sys
import shutil

SERVICE_NAME_TEMPLATE = "conbee-ad"


def replace_in_file(file_name,new_name):
    with open(file_name) as f:
        newText=f.read().replace(SERVICE_NAME_TEMPLATE, new_name)

    with open(file_name, "w") as f:
        f.write(newText)


def rename_file(old_file_path,new_name):
    if SERVICE_NAME_TEMPLATE in old_file_path:
        new_path = old_file_path.replace(SERVICE_NAME_TEMPLATE,new_name)
        os.replace(old_file_path,new_path)
        return new_path
    return old_file_path


def rename_files(new_service_name):
    for root, dirs, files in os.walk("../"+new_service_name):
        for filename in files:
            full_path = root+"/"+filename
            full_path = rename_file(full_path,new_service_name)
            print(full_path)
            replace_in_file(full_path,new_service_name)


if __name__ == "__main__":
    new_service_name = sys.argv[1]
    # debian1 or debian2
    shutil.copytree("./", "../"+new_service_name, ignore=shutil.ignore_patterns('.idea'))
    rename_files(new_service_name)