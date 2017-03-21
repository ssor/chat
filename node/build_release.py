import os
import shutil


if os.path.exists("release") == False:
	os.mkdir("release")

# code = os.system("go build -o ./release/chatserver_dawin main.go")
code = os.system("GOOS=linux GOARCH=amd64 go build -o ./release/chatserver_linux64 main.go")
if code > 0:
	raise ValueError("build chatserver error")

dirSrc = "./"
dirDest = "./release"

def getDestPath(path, src, dest):
	# print("src:  ", src)
	# print("dest: ", dest)
	if src == "./":
		if path.find("./") == 0:
			path = path[2:]
		# print(path)
		return os.path.join(dest, path)

	# print("just replace")
	return path.replace(src, dest)

def hasHidePath(path):
	paths = path.split(os.sep)
	# paths = os.path.split(path)
	# print("paths: ", paths)
	if paths[0] == ".":
		paths = paths[1:]

	if len(paths[0]) <= 0:
		# print(path, " does NOT have hide path")
		return False

	for p in paths:
		# print("p => ",p)
		if p[0] == ".":
			# print(path, " does have hide path")
			return True

	# print(path, " does NOT have hide path")
	return False

# set files and paths that should be exclued from release files
ignoredDirs = ["controllers", "libs", "models", "routers", "tests", "release", "test", "chatFiles"]
ignoredFiles = ["build_release.py", "main.go", "conf/config.json"]


ignoredPaths = []

for d in ignoredDirs:
	ignoredPaths.append(os.path.join(dirSrc, d))

for f in ignoredFiles:
	ignoredPaths.append(os.path.join(dirSrc, f))


# print("ignoredPaths:")
# print(ignoredPaths)

def sha1OfFile(filepath):
    import hashlib
    sha = hashlib.sha1()
    with open(filepath, 'rb') as f:
        while True:
            block = f.read(2**10) # Magic number: one-megabyte blocks.
            if not block: break
            sha.update(block)
        return sha.hexdigest()

def isPathIgnored(filepath):
	for p in ignoredPaths:
		index = filepath.find(p)
		if index == 0:
			# print("--- ignored ! : ", filepath)
			return True

	# print("+++ Not ignored ! : ", filepath)
	return False

def buildRelease():
	print("bulid release tool start...")

	for root, dirs, files in os.walk(dirSrc):
		# print("root: ", root)
		if hasHidePath(root):
			continue

		# print("dirs: ", dirs)
		# print("files:", files)

		for x in dirs:
			srcPath = os.path.join(root,x)
			# destRelativePath = srcPath.replace(dirSrc, dirDest)
			destRelativePath = getDestPath(srcPath, dirSrc, dirDest)
			# print("srcPath: ", srcPath)
			# print("destRelativePath: ", destRelativePath)

			if hasHidePath(x):
				continue

			if isPathIgnored(srcPath) == True:
				# print(srcPath, "ignored !!!")
				continue

			if os.path.exists(destRelativePath) == False:
				# print(destRelativePath, " dir does not exists")
				try:
					os.mkdir(destRelativePath)
					print("duplicate dir ", srcPath," OK")
				except OSError as e:
					print("mkdir error: ", e)
					raise e


		for x in files:
			filePath = os.path.join(root,x)
			# destRelativePath = filePath.replace(dirSrc, dirDest)
			destRelativePath = getDestPath(filePath, dirSrc, dirDest)
			# print("filePath", filePath)
			# print("destRelativePath: ", destRelativePath)

			if hasHidePath(x):
				continue

			if isPathIgnored(filePath) == True:
				continue

			if os.path.exists(destRelativePath) == False:
				# print(destRelativePath, " file does not exists")
				try:
					shutil.copy(filePath, destRelativePath)
					print("copy file ", filePath, " OK")
				except Exception as e:
					print("copy file error: ", filePath, " ", e)
					raise e
			else:
				destSha1 = sha1OfFile(destRelativePath)
				srcSha1 = sha1OfFile(filePath)
				if destSha1 != srcSha1:
					# print("file changed ", filePath)
					try:
						os.remove(destRelativePath)
						shutil.copy(filePath, destRelativePath)
						print("copy file ", filePath, " OK")
					except Exception as e:
						raise e

	print("build OK")	
	print("----------------------------------------------------------------")

buildRelease()
