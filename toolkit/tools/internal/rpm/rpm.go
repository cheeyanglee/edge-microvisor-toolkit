// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rpm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/microsoft/azurelinux/toolkit/tools/internal/exe"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/file"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/logger"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/shell"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/sliceutils"
)

const (
	// TargetArgument specifies to build for a target platform (i.e., aarch64-mariner-linux)
	TargetArgument = "--target"

	// BuildRequiresArgument specifies the build requires argument to be used with rpm tools
	BuildRequiresArgument = "--buildrequires"

	// QueryHeaderArgument specifies the srpm argument to be used with rpm tools
	QueryHeaderArgument = "--srpm"

	// QueryBuiltRPMHeadersArgument specifies that only rpm packages that would be built from a given spec should be queried.
	QueryBuiltRPMHeadersArgument = "--builtrpms"

	// DistTagDefine specifies the dist tag option for rpm tool commands
	DistTagDefine = "dist"

	// DistroReleaseVersionDefine specifies the distro release version option for rpm tool commands
	DistroReleaseVersionDefine = "distro_release_version"

	// DistroBuildNumberDefine specifies the distro build number option for rpm tool commands
	DistroBuildNumberDefine = "mariner_build_number"

	// SourceDirDefine specifies the source directory option for rpm tool commands
	SourceDirDefine = "_sourcedir"

	// TopDirDefine specifies the top directory option for rpm tool commands
	TopDirDefine = "_topdir"

	// WithCheckDefine specifies the with_check option for rpm tool commands
	WithCheckDefine = "with_check"

	// NoCompatibleArchError specifies the error message when processing a SPEC written for a different architecture.
	NoCompatibleArchError = "error: No compatible architectures found for build"

	// Azure LinuxModuleLdflagsDefine specifies the variable used to enable linking ELF binaries with module_info.ld metadata.
	AzureLinuxModuleLdflagsDefine = "distro_module_ldflags "

	// Azure LinuxCCacheDefine enables ccache in the Azure Linux build system
	AzureLinuxCCacheDefine = "ccache_enabled"

	// MaxCPUDefine specifies the max number of CPUs to use for parallel build
	MaxCPUDefine = "_smp_ncpus_max"
)

const (
	packageFQNRegexMatchSubString  = iota
	packageFQNRegexNameIndex       = iota
	packageFQNRegexEpochIndex      = iota
	packageFQNRegexVersionIndex    = iota
	packageFQNRegexReleaseIndex    = iota
	packageFQNRegexArchIndex       = iota
	packageFQNRegexExtensionIndex  = iota
	packageFQNRegexExpectedMatches = iota
)

const (
	installedRPMRegexMatchSubString  = iota
	installedRPMRegexRPMIndex        = iota
	installedRPMRegexVersionIndex    = iota
	installedRPMRegexArchIndex       = iota
	installedRPMRegexExpectedMatches = iota
)

const (
	rpmProgram      = "rpm"
	rpmSpecProgram  = "rpmspec"
	rpmBuildProgram = "rpmbuild"
)

var (
	goArchToRpmArch = map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}

	// checkSectionRegex is used to determine if a SPEC file has a '%check' section.
	// It works multi-line strings containing the whole file content, thus the need for the 'm' flag.
	checkSectionRegex = regexp.MustCompile(`(?m)^\s*%check`)

	// A full qualified RPM name contains the package name, epoch, version, release, architecture, and extension.
	// Optional fields:
	// 	- epoch,
	// 	- architecture.
	//	- "rpm" extension.
	//
	// Sample match:
	//
	//	pkg-name-0:1.2.3-4.azl3.x86_64.rpm
	//
	// Groups can be used to split it into:
	//   - name:			pkg-name
	//   - epoch:			0
	//   - version:			1.2.3
	//   - release:			4.azl3
	//   - architecture:	x86_64
	//   - extension:		rpm
	packageFQNRegex = regexp.MustCompile(`^\s*(\S+[^-])-(?:(\d+):)?(\d[^-:_]*)-(\d+(?:[^-\s]*?))(?:\.(noarch|x86_64|aarch64|src))?(?:\.(rpm))?\s*$`)

	// Output from 'rpm' prints installed RPMs in a line with the following format:
	//
	//	D: ========== +++ [name]-([epoch]:)[version]-[release].[distribution] [architecture]-linux [hex_value]
	//
	// Example:
	//
	//	D: ========== +++ systemd-devel-239-42.azl3 x86_64-linux 0x0
	installedRPMRegex = regexp.MustCompile(`^D: =+ \+{3} (\S+)-([^-]+-[^-]+) (\S+)-linux.*$`)

	// For most use-cases, the distro name abbreviation and major version are set by the exe package. However, if the
	// module is used outside of the main Azure Linux build system, the caller can override these values with SetDistroMacros().
	distNameAbreviation, distMajorVersion = loadLdDistroFlags()
)

// checkDistroMacros validates the distro macro values.
func checkDistroMacros(nameAbreviation string, majorVersion int) error {
	if majorVersion < 1 || nameAbreviation == "" {
		err := fmt.Errorf("failed to set distro defines (%s, %d), invalid name or version", nameAbreviation, majorVersion)
		return err
	}
	return nil
}

// loadDistroFlags will load the values of exe.DistroNameAbbreviation and exe.DistroMajorVersion into the local copies
// after validating them.
func loadLdDistroFlags() (string, int) {
	version, err := strconv.Atoi(exe.DistroMajorVersion)
	if err != nil {
		err = fmt.Errorf("failed to convert distro major version (%s) to int:\n%w", exe.DistroMajorVersion, err)
		panic(err)
	}

	err = checkDistroMacros(exe.DistroNameAbbreviation, version)
	if err != nil {
		err = fmt.Errorf("failed to load distro defines from exe package:\n%w", err)
		panic(err)
	}
	return exe.DistroNameAbbreviation, version
}

// GetRpmArch converts the GOARCH arch into an RPM arch
func GetRpmArch(goArch string) (rpmArch string, err error) {
	rpmArch, ok := goArchToRpmArch[goArch]
	if !ok {
		err = fmt.Errorf("unknown GOARCH detected (%s)", goArch)
	}
	return
}

func GetBasePackageNameFromSpecFile(specPath string) (basePackageName string, err error) {

	baseName := filepath.Base(specPath)
	if baseName == "" {
		return "", fmt.Errorf("failed to extract file name from specPath (%s)", specPath)
	}

	fileExtension := filepath.Ext(baseName)
	if fileExtension == "" {
		return "", fmt.Errorf("failed to extract file extension from file name (%s)", baseName)
	}

	basePackageName = baseName[:len(baseName)-len(fileExtension)]

	return
}

func GetMacroDir() (macroDir string, err error) {
	return getMacroDirWithFallback(false)
}

// Queries rpm for the current macro directory via --eval %_rpmmacrodir
func getMacroDirWithFallback(allowDefault bool) (macroDir string, err error) {
	const (
		macro         = "%_rpmmacrodir"
		defaultRpmDir = "/usr/lib/rpm/macros.d"
	)

	// This should continue to work even if the rpm command is not available (ie unit tests).
	rpmFound, err := file.CommandExists(rpmProgram)
	if err != nil {
		return "", fmt.Errorf("failed to check if rpm is installed:\n%w", err)
	}
	if !rpmFound {
		if allowDefault {
			return defaultRpmDir, nil
		} else {
			return "", fmt.Errorf("rpm is not installed, can't query for macro directory")
		}
	}

	lines, err := executeRpmCommand(rpmProgram, "--eval", macro)
	if err != nil {
		return "", fmt.Errorf("failed to get macro directory:\n%w", err)
	}
	if len(lines) != 1 {
		return "", fmt.Errorf("unexpected output from 'rpm --eval %s': '%v'", macro, lines)
	}
	return lines[0], nil
}

// ExtractNameFromRPMPath strips the version from an RPM file name. i.e. pkg-name-1.2.3-4.cm2.x86_64.rpm -> pkg-name
func ExtractNameFromRPMPath(rpmFilePath string) (packageName string, err error) {
	baseName := filepath.Base(rpmFilePath)

	matches := packageFQNRegex.FindStringSubmatch(baseName)

	// If the path is invalid, return empty string. We consider any string that has at least 1 '-' characters valid.
	if matches == nil {
		err = fmt.Errorf("invalid RPM file path (%s), can't extract name", rpmFilePath)
		return
	}

	return matches[packageFQNRegexNameIndex], nil
}

// getCommonBuildArgs will generate arguments to pass to 'rpmbuild'.
func getCommonBuildArgs(outArch, srpmFile string, defines map[string]string) (buildArgs []string, err error) {
	const (
		os          = "linux"
		queryFormat = ""
		vendor      = "Edge Microvisor Toolkit "
	)

	buildArgs = []string{"--nodeps"}

	// buildArch is the arch of the build machine
	// outArch is the arch of the machine that will run the resulting binary
	buildArch, err := GetRpmArch(runtime.GOARCH)
	if err != nil {
		return
	}

	if buildArch != outArch && outArch != "noarch" {
		tuple := outArch + "-" + vendor + "-" + os
		logger.Log.Debugf("Applying RPM target tuple (%s)", tuple)
		buildArgs = append(buildArgs, TargetArgument, tuple)
	}

	return formatCommandArgs(buildArgs, srpmFile, queryFormat, defines), nil
}

// sanitizeOutput will take the raw output from an RPM command and split by new line,
// trimming whitespace and removing blank lines.
func sanitizeOutput(rawResults string) (sanitizedOutput []string) {
	rawSplitOutput := strings.Split(rawResults, "\n")

	for _, line := range rawSplitOutput {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		sanitizedOutput = append(sanitizedOutput, trimmedLine)
	}

	return
}

// formatCommand will generate an RPM command to execute.
func formatCommandArgs(extraArgs []string, file, queryFormat string, defines map[string]string) (commandArgs []string) {
	commandArgs = append(commandArgs, extraArgs...)

	if queryFormat != "" {
		commandArgs = append(commandArgs, "--qf", queryFormat)
	}

	for k, v := range defines {
		commandArgs = append(commandArgs, "-D", fmt.Sprintf(`%s %s`, k, v))
	}

	commandArgs = append(commandArgs, file)

	return
}

// executeRpmCommand will execute an RPM command and return its output split
// by new line and whitespace trimmed.
func executeRpmCommand(program string, args ...string) (results []string, err error) {
	stdout, err := executeRpmCommandRaw(program, args...)

	return sanitizeOutput(stdout), err
}

// executeRpmCommandRaw will execute an RPM command and return stdout in form of unmodified strings.
func executeRpmCommandRaw(program string, args ...string) (stdout string, err error) {
	stdout, stderr, err := shell.Execute(program, args...)
	if err != nil {
		// When dealing with a SPEC/package intended for a different architecture, explicitly set the error message
		// to a known value so the invoker can check for it.
		//
		// All other errors will be treated normally.
		if strings.Contains(stderr, NoCompatibleArchError) {
			logger.Log.Debug(stderr)
			err = fmt.Errorf(NoCompatibleArchError)
		} else {
			logger.Log.Warn(stderr)
		}
	}

	return
}

// DefaultDistroDefines returns a new map of default defines that can be used during RPM queries that also includes
// the distro tags like '%dist', '%azl'.
func DefaultDistroDefines(runChecks bool, distTag string) map[string]string {
	defines := defaultDefines(runChecks)
	defines[DistTagDefine] = distTag
	defines[distNameAbreviation] = fmt.Sprintf("%d", distMajorVersion)
	return defines
}

// DisableBuildRequiresDefines sets the macro to disable documentation files when installing RPMs.
// - defines: optional map of defines to update. If nil, a new map will be created.
func DisableDocumentationDefines() map[string]string {
	return map[string]string{
		"_excludedocs": "1",
	}
}

// OverrideLocaleDefines sets the macro to override the default locales when installing RPMs.
// - defines: optional map of defines to update. If nil, a new map will be created.
// - overrideLocale: the locale string to set as the default. Should be of the form ""
func OverrideLocaleDefines(overrideLocale string) map[string]string {
	return map[string]string{
		"_install_langs": overrideLocale,
	}
}

// DefaultDefines returns a new map of default defines that can be used during RPM queries.
func defaultDefines(runCheck bool) map[string]string {
	// "with_check" definition should align with the RUN_CHECK Make variable whenever possible
	withCheckSetting := "0"
	if runCheck {
		withCheckSetting = "1"
	}

	return map[string]string{
		WithCheckDefine: withCheckSetting,
	}
}

// GetInstalledPackages returns a string list of all packages installed on the system
// in the "[name]-[version]-[release].[distribution].[architecture]" format.
// Example: tdnf-2.1.0-4.azl3.x86_64
func GetInstalledPackages() (result []string, err error) {
	const queryArg = "-qa"

	return executeRpmCommand(rpmProgram, queryArg)
}

// QuerySPEC queries a SPEC file with queryFormat. Returns the output split by line and trimmed.
func QuerySPEC(specFile, sourceDir, queryFormat, arch string, defines map[string]string, extraArgs ...string) (result []string, err error) {
	const queryArg = "-q"

	extraArgs = append(extraArgs, queryArg)

	// Apply --target arch argument
	extraArgs = append(extraArgs, TargetArgument, arch)

	allDefines := updateSourceDirDefines(defines, sourceDir)

	args := formatCommandArgs(extraArgs, specFile, queryFormat, allDefines)
	return executeRpmCommand(rpmSpecProgram, args...)
}

// QuerySPECForBuiltRPMs queries a SPEC file with queryFormat. Returns only the subpackages, which generate a .rpm file.
func QuerySPECForBuiltRPMs(specFile, sourceDir, arch string, defines map[string]string) (result []string, err error) {
	const queryFormat = "%{nevra}\n"

	return QuerySPEC(specFile, sourceDir, queryFormat, arch, defines, QueryBuiltRPMHeadersArgument)
}

// QueryPackage queries an RPM or SRPM file with queryFormat. Returns the output split by line and trimmed.
func QueryPackage(packageFile, queryFormat string, defines map[string]string, extraArgs ...string) (result []string, err error) {
	const queryArg = "-q"

	extraArgs = append(extraArgs, queryArg)
	args := formatCommandArgs(extraArgs, packageFile, queryFormat, defines)

	return executeRpmCommand(rpmProgram, args...)
}

// QueryPackageFiles queries an RPM for its file contents. The results are split into several categories:
// - allFilesAndDirectories: all files and directories in the package
// - files: all files in the package (ie allFilesAndDirectories minus directories)
// - directories: all directories in the package (ie allFilesAndDirectories minus files, symlinks etc.)
// - documentFiles: all files marked as documentation (%doc)
// - licenseFiles: all files marked as license (%license)
func QueryPackageFiles(packageFile string, defines map[string]string,
) (allFilesAndDirectories, files, directories, documentFiles, licenseFiles []string, err error) {
	const allFilesQueryFormat = "[%{FILEMODES:perms} %{FILENAMES}\n]"
	allFilesWithPerms, err := QueryPackage(packageFile, allFilesQueryFormat, defines)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to query package (%s) files:\n%w", packageFile, err)
	}
	// Parse the output of the query to separarate directories. Output will be of the form:
	// 	drwxr-xr-x /a/directory
	// 	-rw-r--r-- /a/directory/a_file
	// Any line that starts with a 'd' is a directory, everything else is a file (or symlink etc.).
	for _, fileLine := range allFilesWithPerms {
		perms, filePath, found := strings.Cut(fileLine, " ")
		if !found {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to parse package (%s) file contents (%s)", packageFile, fileLine)
		}
		if strings.HasPrefix(perms, "d") {
			directories = append(directories, filePath)
		} else {
			files = append(files, filePath)
		}
		allFilesAndDirectories = append(allFilesAndDirectories, filePath)
	}

	// rpm has dedicated tags for documentation and license files, so we can query them directly.
	documentFiles, err = QueryPackage(packageFile, "", defines, "-d")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to query package (%s) documentation files:\n%w", packageFile, err)
	}

	licenseFiles, err = QueryPackage(packageFile, "", defines, "-L")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to query package (%s) license files:\n%w", packageFile, err)
	}

	return allFilesAndDirectories, files, directories, documentFiles, licenseFiles, nil
}

// BuildRPMFromSRPM builds an RPM from the given SRPM file but does not run its '%check' section.
func BuildRPMFromSRPM(srpmFile, outArch string, defines map[string]string) (err error) {
	const squashErrors = true

	commonBuildArgs, err := getCommonBuildArgs(outArch, srpmFile, defines)
	if err != nil {
		return
	}

	args := []string{"--nocheck", "--rebuild"}
	args = append(args, commonBuildArgs...)

	return shell.ExecuteLive(squashErrors, rpmBuildProgram, args...)
}

// GenerateSRPMFromSPEC generates an SRPM for the given SPEC file
func GenerateSRPMFromSPEC(specFile, topDir string, defines map[string]string) (err error) {
	const (
		generateSRPMArg = "-bs"
		queryFormat     = ""
	)

	var allDefines map[string]string
	extraArgs := []string{generateSRPMArg}

	if topDir == "" {
		allDefines = defines
	} else {
		allDefines = make(map[string]string)
		for k, v := range defines {
			allDefines[k] = v
		}

		allDefines[TopDirDefine] = topDir
	}

	args := formatCommandArgs(extraArgs, specFile, queryFormat, allDefines)
	_, stderr, err := shell.Execute(rpmBuildProgram, args...)
	if err != nil {
		err = fmt.Errorf("%v\n%w", stderr, err)
	}

	return
}

// InstallRPM installs the given RPM or SRPM
func InstallRPM(rpmFile string) (err error) {
	const installOption = "-ihv"

	logger.Log.Debugf("Installing RPM (%s)", rpmFile)

	_, stderr, err := shell.Execute(rpmProgram, installOption, rpmFile)
	if err != nil {
		err = fmt.Errorf("%v\n%w", stderr, err)
	}

	return
}

// QueryRPMProvides returns what an RPM file provides.
// This includes any provides made by a generator and files provided by the rpm.
func QueryRPMProvides(rpmFile string) (provides []string, err error) {
	const queryProvidesOption = "-qlPp"

	logger.Log.Debugf("Querying RPM provides (%s)", rpmFile)
	stdout, stderr, err := shell.Execute(rpmProgram, queryProvidesOption, rpmFile)
	if err != nil {
		err = fmt.Errorf("%v\n%w", stderr, err)
		return
	}

	provides = sanitizeOutput(stdout)
	return
}

// ResolveCompetingPackages takes in a list of RPMs and returns only the ones, which would
// end up being installed after resolving outdated, obsoleted, or conflicting packages.
func ResolveCompetingPackages(rootDir string, rpmPaths ...string) (resolvedRPMs []string, err error) {
	args := []string{
		"-Uvvh",
		"--replacepkgs",
		"--nodeps",
		"--root",
		rootDir,
		"--test",
	}
	args = append(args, rpmPaths...)

	// Output of interest is printed to stderr.
	_, stderr, err := shell.Execute(rpmProgram, args...)
	if err != nil {
		err = fmt.Errorf("%v\n%w", stderr, err)
		return
	}

	splitStdout := strings.Split(stderr, "\n")
	uniqueResolvedRPMs := map[string]bool{}
	for _, line := range splitStdout {
		if match, rpmName := extractCompetingPackageInfoFromLine(line); match {
			uniqueResolvedRPMs[rpmName] = true
		}
	}

	resolvedRPMs = sliceutils.SetToSlice(uniqueResolvedRPMs)
	return
}

// QueryRPMSpecName returns an RPM spec name.
func QueryRPMSpecName(rpmFile string) (specName string, err error) {
	const query = "-qp"
	const format = "--qf"
	const querystring = "%{SOURCERPM}\t-%{VERSION}\n"

	cleanRpm := rpmFile + ".rpm"

	logger.Log.Debugf("Querying RPM Spec name (%v)", rpmFile)
	stdout, stderr, err := shell.Execute(rpmProgram, query, format, querystring, cleanRpm)
	if err != nil {
		err = fmt.Errorf("%v\n%w", stderr, err)
		logger.Log.Warnf("Error Encounter for RPM Query package %v: %v", cleanRpm, stdout)
		return
	}

	// get name
	// Split the result based on tabs
	parts := strings.Split(stdout, "\t")

	// Check if we have at least two parts
	if len(parts) < 2 {
		logger.Log.Warnf("Error Encounter for RPM Query package %v: %v : no version or sourcerpm", cleanRpm, stdout)
		return
	}

	// Extract the name (first string) and the version (delimiter)
	firstString := strings.TrimSpace(parts[0])
	delimiter := strings.TrimSpace(parts[1])

	// Split the name using the delimiter
	result := strings.Split(firstString, delimiter)

	// return name
	specName = strings.TrimSpace(result[0]) + ".spec"

	return
}

// FindDistributionInSpec searches for a specific .spec file in the given directory and its subdirectories,
// reads it, and extracts the Distribution value.
func FindDistributionInSpec(dir, specFileName string) (string, error) {
	var distribution string

	// Walk through the directory and its subdirectories
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file matches the specific .spec file
		if !info.IsDir() && info.Name() == specFileName {
			// Open the .spec file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Scan the file for the Distribution value
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "Distribution:") {
					distribution = strings.TrimSpace(strings.TrimPrefix(line, "Distribution:"))
					return filepath.SkipDir // Stop searching further
				}
			}

			// Check for errors while scanning the file
			if err := scanner.Err(); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipDir {
		return "", err
	}

	if distribution == "" {
		return "", fmt.Errorf("distribution value not found in %s", specFileName)
	}

	return distribution, nil
}

func extractCompetingPackageInfoFromLine(line string) (match bool, pkgName string) {
	matches := installedRPMRegex.FindStringSubmatch(line)
	if len(matches) == installedRPMRegexExpectedMatches {
		pkgName := matches[installedRPMRegexRPMIndex]
		version := matches[installedRPMRegexVersionIndex]
		arch := matches[installedRPMRegexArchIndex]

		return true, fmt.Sprintf("%s-%s.%s", pkgName, version, arch)
	}
	return false, ""
}

// SpecExclusiveArchIsCompatible verifies the "ExclusiveArch" tag is compatible with the current machine's architecture.
func SpecExclusiveArchIsCompatible(specfile, sourcedir, arch string, defines map[string]string) (isCompatible bool, err error) {
	const (
		exclusiveArchIndex = 0
		exclusiveArchQuery = "[%{EXCLUSIVEARCH} ]"
	)

	// Sanity check that this SPEC is meant to be built for the current machine architecture
	queryOutput, err := QuerySPEC(specfile, sourcedir, exclusiveArchQuery, arch, defines, QueryHeaderArgument)
	if err != nil {
		err = fmt.Errorf("failed to query SPEC (%s):\n%w", specfile, err)
		return
	}

	// Empty result means the package is buildable for all architectures.
	if len(queryOutput) == 0 {
		isCompatible = true
		return
	}

	isCompatible = strings.Contains(queryOutput[exclusiveArchIndex], arch)

	return
}

// SpecExcludeArchIsCompatible verifies the "ExcludeArch" tag is compatible with the current machine's architecture.
func SpecExcludeArchIsCompatible(specfile, sourcedir, arch string, defines map[string]string) (isCompatible bool, err error) {
	const (
		excludeArchIndex = 0
		excludeArchQuery = "[%{EXCLUDEARCH} ]"
	)

	queryOutput, err := QuerySPEC(specfile, sourcedir, excludeArchQuery, arch, defines, QueryHeaderArgument)
	if err != nil {
		err = fmt.Errorf("failed to query SPEC (%s):\n%w", specfile, err)
		return
	}

	// Empty result means the package is buildable for all architectures.
	if len(queryOutput) == 0 {
		isCompatible = true
		return
	}

	isCompatible = !strings.Contains(queryOutput[excludeArchIndex], arch)

	return
}

// SpecArchIsCompatible verifies the spec is buildable on the current machine's architecture.
func SpecArchIsCompatible(specfile, sourcedir, arch string, defines map[string]string) (isCompatible bool, err error) {
	isCompatible, err = SpecExclusiveArchIsCompatible(specfile, sourcedir, arch, defines)
	if err != nil {
		return
	}

	if isCompatible {
		return SpecExcludeArchIsCompatible(specfile, sourcedir, arch, defines)
	}

	return
}

// SpecHasCheckSection verifies if the spec has the '%check' section.
func SpecHasCheckSection(specFile, sourceDir, arch string, defines map[string]string) (hasCheckSection bool, err error) {
	const (
		parseSwitch = "--parse"
		queryFormat = ""
	)

	basicArgs := []string{
		parseSwitch,
		TargetArgument,
		arch,
	}
	allDefines := updateSourceDirDefines(defines, sourceDir)
	args := formatCommandArgs(basicArgs, specFile, queryFormat, allDefines)

	stdout, err := executeRpmCommandRaw(rpmSpecProgram, args...)

	return checkSectionRegex.MatchString(stdout), err
}

// BuildCompatibleSpecsList builds a list of spec files in a directory that are compatible with the build arch. Paths
// are relative to the 'baseDir' directory. This function should generally be used from inside a chroot to ensure the
// correct defines are available.
func BuildCompatibleSpecsList(baseDir string, inputSpecPaths []string, defines map[string]string) (filteredSpecPaths []string, err error) {
	var specPaths []string
	if len(inputSpecPaths) > 0 {
		specPaths = inputSpecPaths
	} else {
		specPaths, err = buildAllSpecsList(baseDir)
		if err != nil {
			return
		}
	}

	return filterCompatibleSpecs(specPaths, defines)
}

// StripEpochFromPackageFullQualifiedName removes the epoch from a package full qualified name if it is present.
// Example:
//
//	"pkg-name-0:1.2.3-4.azl3.x86_64" -> "pkg-name-1.2.3-4.azl3.x86_64"
func StripEpochFromPackageFullQualifiedName(packageFQN string) string {
	var packageFQNBuilder strings.Builder

	matches := packageFQNRegex.FindStringSubmatch(packageFQN)
	if matches == nil {
		return packageFQN
	}

	packageFQNBuilder.WriteString(matches[packageFQNRegexNameIndex])
	packageFQNBuilder.WriteString("-")

	packageFQNBuilder.WriteString(matches[packageFQNRegexVersionIndex])
	packageFQNBuilder.WriteString("-")

	packageFQNBuilder.WriteString(matches[packageFQNRegexReleaseIndex])

	if matches[packageFQNRegexArchIndex] != "" {
		packageFQNBuilder.WriteString(".")
		packageFQNBuilder.WriteString(matches[packageFQNRegexArchIndex])
	}

	if matches[packageFQNRegexExtensionIndex] != "" {
		packageFQNBuilder.WriteString(".")
		packageFQNBuilder.WriteString(matches[packageFQNRegexExtensionIndex])
	}

	return packageFQNBuilder.String()
}

// TestRPMFromSRPM builds an RPM from the given SRPM and runs its '%check' section SRPM file
// but it does not generate any RPM packages.
func TestRPMFromSRPM(srpmFile, outArch string, defines map[string]string) (err error) {
	const squashErrors = true

	commonBuildArgs, err := getCommonBuildArgs(outArch, srpmFile, defines)
	if err != nil {
		return
	}

	args := []string{"-ri"}
	args = append(args, commonBuildArgs...)

	return shell.ExecuteLive(squashErrors, rpmBuildProgram, args...)
}

// buildAllSpecsList builds a list of all spec files in the directory. Paths are relative to the base directory.
func buildAllSpecsList(baseDir string) (specPaths []string, err error) {
	specFilesGlob := filepath.Join(baseDir, "**", "*.spec")

	specPaths, err = filepath.Glob(specFilesGlob)
	if err != nil {
		specPaths, err = nil, fmt.Errorf("failed to enumerate all spec files with (%s):\n%w", specFilesGlob, err)
		return
	}

	return
}

// filterCompatibleSpecs filters a list of spec files in the chroot's SPECs directory that are compatible with the build arch.
func filterCompatibleSpecs(inputSpecPaths []string, defines map[string]string) (filteredSpecPaths []string, err error) {
	var specCompatible bool

	buildArch, err := GetRpmArch(runtime.GOARCH)
	if err != nil {
		return
	}

	type specArchResult struct {
		compatible bool
		path       string
		err        error
	}
	resultsChannel := make(chan specArchResult, len(inputSpecPaths))

	for _, specFilePath := range inputSpecPaths {
		specDirPath := filepath.Dir(specFilePath)

		go func(pathIter string) {
			specCompatible, err = SpecArchIsCompatible(pathIter, specDirPath, buildArch, defines)
			if err != nil {
				err = fmt.Errorf("failed while querrying spec (%s). Error: %v.", pathIter, err)
			}
			resultsChannel <- specArchResult{
				compatible: specCompatible,
				path:       pathIter,
				err:        err,
			}
		}(specFilePath)
	}

	for i := 0; i < len(inputSpecPaths); i++ {
		result := <-resultsChannel
		if result.err != nil {
			err = result.err
			return
		}
		if result.compatible {
			filteredSpecPaths = append(filteredSpecPaths, result.path)
		}
	}

	return
}

// updateSourceDirDefines adds the source directory to the defines map if it is not empty.
// To query some SPECs the source directory must be set
// since the SPEC file may use `%include` on a source file.
func updateSourceDirDefines(defines map[string]string, sourceDir string) (updatedDefines map[string]string) {
	updatedDefines = make(map[string]string)
	for key, value := range defines {
		updatedDefines[key] = value
	}

	if sourceDir != "" {
		updatedDefines[SourceDirDefine] = sourceDir
	}

	return
}
